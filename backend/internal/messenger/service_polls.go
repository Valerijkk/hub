package messenger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/poll"
	"github.com/valerii/hub/backend/ent/polloption"
	"github.com/valerii/hub/backend/ent/pollvote"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// CreatePoll создаёт новый опрос в канале.
// Проверяет членство пользователя в канале, создаёт опрос и варианты ответов.
func (s *MessengerService) CreatePoll(ctx context.Context, channelID, creatorID uuid.UUID, req CreatePollRequest) (*PollResponse, error) {
	if s.db == nil {
		return nil, apperrors.NewInternal("база данных недоступна")
	}

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

	// Создаём опрос в транзакции.
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка начала транзакции")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	p, err := tx.Poll.Create().
		SetChannelID(channelID).
		SetCreatorID(creatorID).
		SetQuestion(req.Question).
		SetIsAnonymous(req.IsAnonymous).
		SetIsMultiChoice(req.IsMultiChoice).
		Save(ctx)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания опроса")
	}

	// Создаём варианты ответов.
	options := make([]*ent.PollOption, 0, len(req.Options))
	for i, optText := range req.Options {
		opt, optErr := tx.PollOption.Create().
			SetPollID(p.ID).
			SetText(optText).
			SetPosition(i).
			Save(ctx)
		if optErr != nil {
			err = optErr
			return nil, apperrors.NewInternal("ошибка создания варианта ответа")
		}
		options = append(options, opt)
	}

	if err = tx.Commit(); err != nil {
		return nil, apperrors.NewInternal("ошибка фиксации транзакции")
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
	msgID, _ := s.repo.InsertMessage(ctx, pollDoc)
	if msgID != "" {
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
	if s.db == nil {
		return nil, apperrors.NewInternal("база данных недоступна")
	}

	// Получаем опрос.
	p, err := s.db.Poll.Query().
		Where(poll.IDEQ(pollID)).
		WithOptions(func(q *ent.PollOptionQuery) {
			q.Order(ent.Asc(polloption.FieldPosition))
		}).
		Only(ctx)
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
	validOptions := make(map[uuid.UUID]bool, len(p.Edges.Options))
	for _, opt := range p.Edges.Options {
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

	// Транзакция: удаляем старые голоса, добавляем новые.
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка начала транзакции")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Удаляем предыдущие голоса пользователя в этом опросе.
	_, err = tx.PollVote.Delete().
		Where(
			pollvote.PollIDEQ(pollID),
			pollvote.UserIDEQ(userID),
		).
		Exec(ctx)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка удаления предыдущих голосов")
	}

	// Создаём новые голоса.
	for _, oid := range optionIDs {
		_, voteErr := tx.PollVote.Create().
			SetPollID(pollID).
			SetOptionID(oid).
			SetUserID(userID).
			Save(ctx)
		if voteErr != nil {
			err = voteErr
			return nil, apperrors.NewInternal("ошибка сохранения голоса")
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, apperrors.NewInternal("ошибка фиксации транзакции")
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
	if s.db == nil {
		return nil, apperrors.NewInternal("база данных недоступна")
	}

	p, err := s.db.Poll.Get(ctx, pollID)
	if err != nil {
		return nil, apperrors.NewNotFound("опрос не найден")
	}

	if p.CreatorID != userID {
		return nil, apperrors.NewForbidden("только создатель может закрыть опрос")
	}

	if p.IsClosed {
		return nil, apperrors.NewBadRequest("опрос уже закрыт")
	}

	_, err = s.db.Poll.UpdateOneID(pollID).
		SetIsClosed(true).
		Save(ctx)
	if err != nil {
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
	if s.db == nil {
		return nil, apperrors.NewInternal("база данных недоступна")
	}

	p, err := s.db.Poll.Query().
		Where(poll.IDEQ(pollID)).
		WithOptions(func(q *ent.PollOptionQuery) {
			q.Order(ent.Asc(polloption.FieldPosition))
		}).
		Only(ctx)
	if err != nil {
		return nil, apperrors.NewNotFound("опрос не найден")
	}

	// Получаем все голоса для этого опроса.
	votes, err := s.db.PollVote.Query().
		Where(pollvote.PollIDEQ(pollID)).
		All(ctx)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения голосов")
	}

	return toPollResponse(p, p.Edges.Options, votes, currentUserID, p.IsAnonymous), nil
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
