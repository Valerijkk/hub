package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// CreatePoll создаёт новый опрос в канале.
// Проверяет членство пользователя в канале, создаёт опрос и варианты ответов.
func (s *MessengerService) CreatePoll(ctx context.Context, channelID, creatorID uuid.UUID, req CreatePollRequest) (*PollResponse, error) {
	// Проверяем членство в канале.
	isMember, err := s.repo.IsMember(ctx, channelID, creatorID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	// Валидация количества вариантов.
	if len(req.Options) < 2 {
		return nil, apperrors.NewBadRequest("необходимо минимум 2 варианта ответа")
	}
	if len(req.Options) > 10 {
		return nil, apperrors.NewBadRequest("максимум 10 вариантов ответа")
	}
	if req.Question == "" {
		return nil, apperrors.NewBadRequest("вопрос не может быть пустым")
	}

	// Создаём опрос в транзакции — Tx инкапсулирован внутри репозитория.
	p, options, err := s.repo.CreatePollWithOptions(ctx, CreatePollInput{
		ChannelID:     channelID,
		CreatorID:     creatorID,
		Question:      req.Question,
		IsAnonymous:   req.IsAnonymous,
		IsMultiChoice: req.IsMultiChoice,
		Options:       req.Options,
	})
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания опроса")
	}

	// Формируем ответ.
	resp := toPollResponse(p, options, nil, creatorID, false)

	// Создаём сообщение-опрос в канале (MongoDB), чтобы опрос отображался в ленте.
	pollDoc := &MessageDocument{
		ChannelID:   channelID,
		SenderID:    creatorID,
		MessageType: "poll",
		PollID:      p.ID.String(),
		ContentJSON: map[string]interface{}{
			"type":     "doc",
			"children": []interface{}{map[string]interface{}{"type": "p", "children": []interface{}{map[string]interface{}{"text": req.Question}}}},
		},
		ContentText: req.Question,
		CreatedAt:   time.Now(),
	}
	// ME-4: раньше ошибка InsertMessage проглатывалась через `msgID, _ := ...`.
	// Это non-critical path (опрос уже создан в Postgres транзакции), но молчать
	// нельзя — без сообщения-обёртки опрос не появится в ленте, и пользователь
	// видит «пропавший» опрос. Логируем WARN и публикуем NATS-событие только при
	// успешной вставке; общий результат CreatePoll не валим, чтобы основной
	// эффект (создание опроса) был виден через GetPollResults.
	msgID, msgErr := s.repo.InsertMessage(ctx, pollDoc)
	switch {
	case msgErr != nil:
		slog.Warn("ошибка вставки сообщения-обёртки опроса",
			"poll_id", p.ID, "channel_id", channelID, "error", msgErr)
	case msgID != "":
		pollDoc.ID = msgID
		// Публикуем message_sent для real-time доставки.
		s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, channelID.String()), MessageSentEvent{
			MessageID:   msgID,
			ChannelID:   channelID,
			SenderID:    creatorID,
			MessageType: "poll",
			PollID:      p.ID.String(),
			ContentText: req.Question,
			CreatedAt:   pollDoc.CreatedAt,
		})
	}

	// Публикуем NATS-событие опроса.
	s.nats.PublishJSON(fmt.Sprintf(SubjectPollCreated, channelID.String()), PollCreatedEvent{
		PollID:    p.ID.String(),
		ChannelID: channelID,
		CreatorID: creatorID,
		Question:  req.Question,
		CreatedAt: p.CreatedAt,
	})

	return resp, nil
}

// VotePoll обрабатывает голосование пользователя в опросе.
// Поддерживает одиночный и множественный выбор. Переголосование удаляет старые голоса.
func (s *MessengerService) VotePoll(ctx context.Context, pollID, userID uuid.UUID, req VotePollRequest) (*PollResponse, error) {
	// Получаем опрос с вариантами.
	p, options, err := s.repo.GetPollWithOptions(ctx, pollID)
	if err != nil {
		return nil, apperrors.NewNotFound("опрос не найден")
	}

	if p.IsClosed {
		return nil, apperrors.NewBadRequest("опрос закрыт")
	}

	// Проверяем членство в канале.
	isMember, err := s.repo.IsMember(ctx, p.ChannelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	// Валидируем option_ids.
	optionIDs := make([]uuid.UUID, 0, len(req.OptionIDs))
	validOptions := make(map[uuid.UUID]bool, len(options))
	for _, opt := range options {
		validOptions[opt.ID] = true
	}
	for _, oidStr := range req.OptionIDs {
		oid, parseErr := uuid.Parse(oidStr)
		if parseErr != nil {
			return nil, apperrors.NewBadRequest("некорректный ID варианта")
		}
		if !validOptions[oid] {
			return nil, apperrors.NewBadRequest("вариант не принадлежит этому опросу")
		}
		optionIDs = append(optionIDs, oid)
	}

	if !p.IsMultiChoice && len(optionIDs) > 1 {
		return nil, apperrors.NewBadRequest("в этом опросе можно выбрать только один вариант")
	}

	// Атомарное переголосование — Tx внутри репозитория.
	if err := s.repo.VotePoll(ctx, pollID, userID, optionIDs); err != nil {
		return nil, apperrors.NewInternal("ошибка сохранения голоса")
	}

	// Получаем обновлённые результаты.
	resp, err := s.GetPollResults(ctx, pollID, userID)
	if err != nil {
		return nil, err
	}

	// Публикуем NATS-событие.
	s.nats.PublishJSON(fmt.Sprintf(SubjectPollVoted, p.ChannelID.String()), PollVotedEvent{
		PollID:    pollID.String(),
		ChannelID: p.ChannelID,
		UserID:    userID,
	})

	return resp, nil
}

// ClosePoll закрывает опрос. Только создатель может закрыть свой опрос.
func (s *MessengerService) ClosePoll(ctx context.Context, pollID, userID uuid.UUID) (*PollResponse, error) {
	p, err := s.repo.GetPoll(ctx, pollID)
	if err != nil {
		return nil, apperrors.NewNotFound("опрос не найден")
	}

	if p.CreatorID != userID {
		return nil, apperrors.NewForbidden("только создатель может закрыть опрос")
	}

	if p.IsClosed {
		return nil, apperrors.NewBadRequest("опрос уже закрыт")
	}

	if err := s.repo.ClosePoll(ctx, pollID); err != nil {
		return nil, apperrors.NewInternal("ошибка закрытия опроса")
	}

	// Получаем обновлённые результаты.
	resp, err := s.GetPollResults(ctx, pollID, userID)
	if err != nil {
		return nil, err
	}

	// Публикуем NATS-событие.
	s.nats.PublishJSON(fmt.Sprintf(SubjectPollClosed, p.ChannelID.String()), PollClosedEvent{
		PollID:    pollID.String(),
		ChannelID: p.ChannelID,
		ClosedBy:  userID,
	})

	return resp, nil
}

// GetPollResults возвращает результаты опроса с учётом текущего пользователя.
func (s *MessengerService) GetPollResults(ctx context.Context, pollID, currentUserID uuid.UUID) (*PollResponse, error) {
	p, options, err := s.repo.GetPollWithOptions(ctx, pollID)
	if err != nil {
		return nil, apperrors.NewNotFound("опрос не найден")
	}

	votes, err := s.repo.ListVotes(ctx, pollID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения голосов")
	}

	return toPollResponse(p, options, votes, currentUserID, p.IsAnonymous), nil
}

// toPollResponse формирует PollResponse из Ent-сущностей.
func toPollResponse(p *ent.Poll, options []*ent.PollOption, votes []*ent.PollVote, currentUserID uuid.UUID, isAnonymous bool) *PollResponse {
	// Группируем голоса по вариантам.
	votesByOption := make(map[uuid.UUID][]uuid.UUID)
	totalVotes := 0
	// Считаем уникальных голосующих (один пользователь = 1 голос даже при multi-choice).
	uniqueVoters := make(map[uuid.UUID]struct{})
	for _, v := range votes {
		votesByOption[v.OptionID] = append(votesByOption[v.OptionID], v.UserID)
		uniqueVoters[v.UserID] = struct{}{}
	}
	totalVotes = len(uniqueVoters)

	optResponses := make([]PollOptionResponse, 0, len(options))
	for _, opt := range options {
		voterUUIDs := votesByOption[opt.ID]
		voted := false
		voterIDs := make([]string, 0, len(voterUUIDs))
		for _, uid := range voterUUIDs {
			if uid == currentUserID {
				voted = true
			}
			if !isAnonymous {
				voterIDs = append(voterIDs, uid.String())
			}
		}

		optResp := PollOptionResponse{
			ID:       opt.ID.String(),
			Text:     opt.Text,
			Position: opt.Position,
			Count:    len(voterUUIDs),
			Voted:    voted,
		}
		if !isAnonymous {
			optResp.VoterIDs = voterIDs
		}
		optResponses = append(optResponses, optResp)
	}

	return &PollResponse{
		ID:            p.ID.String(),
		ChannelID:     p.ChannelID.String(),
		CreatorID:     p.CreatorID.String(),
		Question:      p.Question,
		IsAnonymous:   p.IsAnonymous,
		IsMultiChoice: p.IsMultiChoice,
		IsClosed:      p.IsClosed,
		TotalVotes:    totalVotes,
		Options:       optResponses,
		CreatedAt:     p.CreatedAt,
	}
}
