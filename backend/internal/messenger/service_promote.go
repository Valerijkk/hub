package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Преобразование треда в задачу (Feature #58)
// ---------------------------------------------------------------------------

// IssuePromoter — узкий интерфейс для создания задачи из сообщения треда.
// Реализуется trackermod.TrackerService — определён здесь, чтобы избежать
// обратного импорта messenger -> tracker и сохранить тестируемость через мок.
type promoteSourceLookup interface {
	FindLiveMessageInChannel(
		ctx context.Context,
		channelID uuid.UUID,
		messageID string,
	) (*MessageDocument, bool, error)
}

func lookupPromotionSource(
	ctx context.Context,
	lookup promoteSourceLookup,
	channelID uuid.UUID,
	messageID string,
) (*MessageDocument, error) {
	source, found, err := lookup.FindLiveMessageInChannel(ctx, channelID, messageID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки исходного сообщения")
	}
	if !found {
		return nil, apperrors.NewNotFound("сообщение не найдено")
	}
	if source == nil || source.ID == "" || source.ID != messageID || source.ChannelID != channelID || source.DeletedAt != nil {
		return nil, apperrors.NewInternal("ошибка проверки исходного сообщения")
	}
	return source, nil
}

func lookupPromotionParent(
	ctx context.Context,
	lookup promoteSourceLookup,
	channelID uuid.UUID,
	parentID string,
) (*MessageDocument, error) {
	parent, found, err := lookup.FindLiveMessageInChannel(ctx, channelID, parentID)
	if err != nil {
		return nil, apperrors.NewInternal("\u043e\u0448\u0438\u0431\u043a\u0430 \u043f\u0440\u043e\u0432\u0435\u0440\u043a\u0438 \u0442\u0440\u0435\u0434\u0430")
	}
	if !found {
		return nil, apperrors.NewNotFound("\u0442\u0440\u0435\u0434 \u043d\u0435 \u043d\u0430\u0439\u0434\u0435\u043d")
	}
	if parent == nil || parent.ID == "" || parent.ID != parentID || parent.ChannelID != channelID || parent.DeletedAt != nil {
		return nil, apperrors.NewInternal("\u043e\u0448\u0438\u0431\u043a\u0430 \u043f\u0440\u043e\u0432\u0435\u0440\u043a\u0438 \u0442\u0440\u0435\u0434\u0430")
	}
	if parent.ThreadID != "" {
		return nil, apperrors.NewNotFound("\u0442\u0440\u0435\u0434 \u043d\u0435 \u043d\u0430\u0439\u0434\u0435\u043d")
	}
	return parent, nil
}

type promoteThreadRepliesLookup interface {
	ListLiveThreadRepliesInChannel(
		ctx context.Context,
		channelID uuid.UUID,
		rootID string,
		limit int,
	) ([]*MessageDocument, error)
}

type IssuePromoter interface {
	// CreateIssueFromPromotion создаёт задачу в проекте и возвращает её ключ и id
	// для построения backlink-сообщения.
	CreateIssueFromPromotion(
		ctx context.Context,
		projectID, userID uuid.UUID,
		title, issueType string,
		descriptionJSON map[string]interface{},
	) (issueKey, issueID string, err error)
}

// PromotePermissionChecker — узкий интерфейс для проверки права создавать задачи
// в целевом проекте. Реализуется rbac.RBACService через адаптер в main.go.
//
// Возвращает существование целевого проекта и право пользователя task.create в нём.
type PromotePermissionChecker interface {
	AuthorizePromotionTarget(ctx context.Context, userID, projectID uuid.UUID) (targetFound bool, canCreate bool, err error)
}

// IssueThreadLinker создаёт связь issue↔thread после promote (PROD-03).
type IssueThreadLinker interface {
	LinkThreadToIssue(ctx context.Context, projectID, issueID, userID uuid.UUID, channelID uuid.UUID, threadRootMessageID string) error
}

// SetIssueThreadLinker подключает tracker для автосвязи при promote.
func (s *MessengerService) SetIssueThreadLinker(l IssueThreadLinker) {
	s.promotionConfigMu.Lock()
	defer s.promotionConfigMu.Unlock()
	s.issueThreadLinker = l
}

// SetIssuePromoter подключает реализацию IssuePromoter к сервису.
// Допускается nil — в этом случае /promote-to-issue вернёт 500.
func (s *MessengerService) SetIssuePromoter(p IssuePromoter) {
	s.promotionConfigMu.Lock()
	defer s.promotionConfigMu.Unlock()
	s.issuePromoter = p
}

// SetPromotePermissionChecker подключает RBAC-проверку target-проекта для
// /promote-to-issue. Nil недопустим в prod: PromoteToIssue fail-closed (500).
func (s *MessengerService) SetPromotePermissionChecker(c PromotePermissionChecker) {
	s.promotionConfigMu.Lock()
	defer s.promotionConfigMu.Unlock()
	s.promoteCheck = c
}

// promotionDependencies returns one request-scoped configuration snapshot and
// releases the lock before any external dependency is called.
func (s *MessengerService) promotionDependencies() (IssuePromoter, PromotePermissionChecker, IssueThreadLinker) {
	s.promotionConfigMu.RLock()
	defer s.promotionConfigMu.RUnlock()
	return s.issuePromoter, s.promoteCheck, s.issueThreadLinker
}

// PromoteToIssueRequest — параметры промоушена сообщения в задачу.
type PromoteToIssueRequest struct {
	ProjectID string `json:"project_id" validate:"required"`
	Title     string `json:"title" validate:"required,min=1,max=500"`
	IssueType string `json:"issue_type,omitempty" validate:"omitempty,oneof=bug task story epic sub_task improvement tech_debt spike release"`
}

// PromoteToIssueResponse — результат промоушена.
type PromoteToIssueResponse struct {
	IssueID  string `json:"issue_id"`
	IssueKey string `json:"issue_key"`
}

// checkPromotionTarget maps the target-project authorization result to the
// public PromoteToIssue error contract.
func checkPromotionTarget(
	checker PromotePermissionChecker,
	ctx context.Context,
	userID, projectID uuid.UUID,
) error {
	if checker == nil {
		return apperrors.NewInternal("ошибка проверки прав на создание задачи")
	}
	targetFound, canCreate, err := checker.AuthorizePromotionTarget(ctx, userID, projectID)
	if err != nil {
		return apperrors.NewInternal("ошибка проверки прав на создание задачи")
	}
	if !targetFound && canCreate {
		return apperrors.NewInternal("ошибка проверки прав на создание задачи")
	}
	if !targetFound {
		return apperrors.NewNotFound("проект не найден")
	}
	if !canCreate {
		return apperrors.NewForbidden("нет права создавать задачи в проекте")
	}
	return nil
}

func (s *MessengerService) checkPromotionChannelAccess(
	ctx context.Context,
	userID, channelID uuid.UUID,
) error {
	_, scoped, accessErr := s.resolveChannelAction(ctx, userID, channelID, ChannelActionPromote)
	if accessErr != nil {
		return accessErr
	}
	if scoped {
		return nil
	}
	isMember, membershipErr := s.repo.IsMember(ctx, channelID, userID)
	if membershipErr != nil {
		return apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return apperrors.NewForbidden("вы не являетесь участником канала-источника")
	}
	return nil
}

// PromoteToIssue превращает сообщение треда в задачу трекера и постит
// backlink-сообщение в исходный тред.
//
// Контракт:
//  1. Пользователь должен быть участником канала-источника (IsMember).
//  2. Пользователь должен иметь PermTaskCreate в целевом проекте —
//     проверяет PromotePermissionChecker (fail-closed при nil).
//  3. Описание задачи собирается из родительского сообщения + сиблингов треда:
//     заголовок «Promoted from #channel-name on YYYY-MM-DD», далее markdown
//     список «- @user: msg».
//  4. После успешного создания постится системное сообщение в тред с ссылкой
//     на новую задачу. Ошибка постинга backlink не отменяет промоушн —
//     логируется как warn.
func (s *MessengerService) PromoteToIssue(
	ctx context.Context,
	channelID uuid.UUID,
	messageID string,
	userID uuid.UUID,
	req PromoteToIssueRequest,
) (*PromoteToIssueResponse, error) {
	issuePromoter, promoteChecker, issueThreadLinker := s.promotionDependencies()
	if issuePromoter == nil {
		return nil, apperrors.NewInternal("промоушен задач не сконфигурирован")
	}

	projectID, err := uuid.Parse(strings.TrimSpace(req.ProjectID))
	if err != nil || projectID == uuid.Nil {
		return nil, apperrors.NewBadRequest("некорректный project_id")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, apperrors.NewBadRequest("title обязателен")
	}
	title = truncateRunes(title, 500)

	issueType := strings.TrimSpace(req.IssueType)
	if issueType == "" {
		issueType = "task"
	}

	// 1. Достаём исходное сообщение.
	if accessErr := s.checkPromotionChannelAccess(ctx, userID, channelID); accessErr != nil {
		return nil, accessErr
	}

	sourceLookup, ok := s.repo.(promoteSourceLookup)
	if !ok {
		return nil, apperrors.NewInternal("ошибка проверки исходного сообщения")
	}
	source, sourceErr := lookupPromotionSource(ctx, sourceLookup, channelID, messageID)
	if sourceErr != nil {
		return nil, sourceErr
	}

	// 2b. Снимок проверки цели используется для обеих проверок этого промоушена.
	if checkErr := checkPromotionTarget(promoteChecker, ctx, userID, projectID); checkErr != nil {
		return nil, checkErr
	}

	// 3. Разрешаем корень треда до чтения канала или запуска побочных эффектов.
	parent := source
	parentID := source.ID
	if source.ThreadID != "" {
		resolvedParent, parentErr := lookupPromotionParent(ctx, sourceLookup, source.ChannelID, source.ThreadID)
		if parentErr != nil {
			return nil, parentErr
		}
		parent = resolvedParent
		parentID = parent.ID
	}

	// 4. Читаем только живые ответы того же канала и проверяем целостность набора.
	threadRepliesLookup, ok := s.repo.(promoteThreadRepliesLookup)
	if !ok {
		return nil, apperrors.NewInternal("ошибка проверки треда")
	}
	replies, repliesErr := threadRepliesLookup.ListLiveThreadRepliesInChannel(ctx, source.ChannelID, parentID, 100)
	if repliesErr != nil {
		return nil, apperrors.NewInternal("ошибка проверки треда")
	}
	if len(replies) > 100 {
		return nil, apperrors.NewInternal("ошибка проверки треда")
	}
	replyIDs := make(map[string]struct{}, len(replies))
	sourceReplyCount := 0
	for _, reply := range replies {
		if reply == nil || reply.ID == "" || reply.ChannelID != source.ChannelID || reply.DeletedAt != nil || reply.ThreadID != parentID || reply.ID == parentID {
			return nil, apperrors.NewInternal("ошибка проверки треда")
		}
		if _, exists := replyIDs[reply.ID]; exists {
			return nil, apperrors.NewInternal("ошибка проверки треда")
		}
		replyIDs[reply.ID] = struct{}{}
		if reply.ID == source.ID {
			sourceReplyCount++
		}
	}
	if source.ThreadID != "" && sourceReplyCount == 0 {
		retained := len(replies)
		if retained == 100 {
			retained = 99
		}
		repliesWithSource := make([]*MessageDocument, 0, retained+1)
		repliesWithSource = append(repliesWithSource, replies[:retained]...)
		replies = append(repliesWithSource, source)
	}
	chronologicalReplies := append([]*MessageDocument(nil), replies...)
	sort.Slice(chronologicalReplies, func(i, j int) bool {
		if chronologicalReplies[i].CreatedAt.Equal(chronologicalReplies[j].CreatedAt) {
			return chronologicalReplies[i].ID < chronologicalReplies[j].ID
		}
		return chronologicalReplies[i].CreatedAt.Before(chronologicalReplies[j].CreatedAt)
	})

	// 5. Канал — для имени в описании.
	ch, chErr := s.repo.GetChannel(ctx, source.ChannelID)
	channelName := "channel"
	if chErr == nil && ch != nil && ch.Name != "" {
		channelName = ch.Name
	}

	// 6. Формируем описание (Plate-совместимый JSON: doc → paragraph children).
	description := s.buildPromotionDescription(ctx, channelName, parent, chronologicalReplies)

	// 7. Повторно проверяем источник, корень reply-треда и затем цель непосредственно перед созданием задачи.
	if _, sourceErr := lookupPromotionSource(ctx, sourceLookup, channelID, messageID); sourceErr != nil {
		return nil, sourceErr
	}
	if source.ThreadID != "" {
		if _, parentErr := lookupPromotionParent(ctx, sourceLookup, channelID, source.ThreadID); parentErr != nil {
			return nil, parentErr
		}
	}
	if accessErr := s.checkPromotionChannelAccess(ctx, userID, channelID); accessErr != nil {
		return nil, accessErr
	}
	if checkErr := checkPromotionTarget(promoteChecker, ctx, userID, projectID); checkErr != nil {
		return nil, checkErr
	}
	issueKey, issueID, err := issuePromoter.CreateIssueFromPromotion(ctx, projectID, userID, title, issueType, description)
	if err != nil {
		return nil, err
	}

	// 8. Постим backlink-сообщение в тред. Не валим основной запрос на ошибках.
	if postErr := s.postPromotionBacklink(ctx, source.ChannelID, parentID, userID, issueKey); postErr != nil {
		slog.Warn("promote-to-issue: failed to post backlink message", "error", postErr, "issue_key", issueKey)
	}

	if issueThreadLinker != nil {
		issueUUID, parseErr := uuid.Parse(issueID)
		if parseErr == nil {
			if linkErr := issueThreadLinker.LinkThreadToIssue(ctx, projectID, issueUUID, userID, source.ChannelID, parentID); linkErr != nil {
				slog.Warn("promote-to-issue: failed to link thread", "error", linkErr, "issue_key", issueKey)
			}
		}
	}

	return &PromoteToIssueResponse{IssueID: issueID, IssueKey: issueKey}, nil
}

// buildPromotionDescription собирает Plate-совместимое JSON-описание для задачи
// из родителя треда и его сиблингов.
func (s *MessengerService) buildPromotionDescription(
	ctx context.Context,
	channelName string,
	parent *MessageDocument,
	replies []*MessageDocument,
) map[string]interface{} {
	today := time.Now().UTC().Format("2006-01-02")
	header := fmt.Sprintf("Promoted from #%s on %s", channelName, today)

	// Plate JSON: { type: "doc", children: [ { type: "p", children: [{ text: ... }] }, ... ] }.
	paragraphs := []interface{}{
		mkParagraph(header),
		mkParagraph(""),
		mkParagraph("Thread:"),
	}

	// Если parent есть — добавляем его строкой первым.
	if parent != nil {
		paragraphs = append(paragraphs, mkParagraph(s.formatMessageLine(ctx, parent)))
	}
	for _, m := range replies {
		if m == nil {
			continue
		}
		// Если parent совпал — не дублируем.
		if parent != nil && m.ID == parent.ID {
			continue
		}
		paragraphs = append(paragraphs, mkParagraph(s.formatMessageLine(ctx, m)))
	}

	return map[string]interface{}{
		"type":     "doc",
		"children": paragraphs,
	}
}

// formatMessageLine форматирует одну строку треда: "- @display: text".
func (s *MessengerService) formatMessageLine(ctx context.Context, m *MessageDocument) string {
	name := "user"
	if m.SenderID != uuid.Nil {
		if displayName, err := s.repo.GetUserDisplayName(ctx, m.SenderID); err == nil && displayName != "" {
			name = displayName
		}
	}
	text := strings.TrimSpace(m.ContentText)
	if text == "" {
		text = "(empty)"
	}
	return fmt.Sprintf("- @%s: %s", name, text)
}

// mkParagraph создаёт один параграф Plate JSON со строкой текста.
func mkParagraph(text string) map[string]interface{} {
	return map[string]interface{}{
		"type": "p",
		"children": []interface{}{
			map[string]interface{}{"text": text},
		},
	}
}

// postPromotionBacklink публикует системное сообщение в треде со ссылкой на задачу.
// Не использует SendMessage (там есть проверка readonly/мембершипа, которые
// для системного backlink не нужны — мы уже всё проверили выше) — пишет
// напрямую через репозиторий и публикует NATS-событие.
func (s *MessengerService) postPromotionBacklink(
	ctx context.Context,
	channelID uuid.UUID,
	threadID string,
	userID uuid.UUID,
	issueKey string,
) error {
	text := fmt.Sprintf("📌 Промоутнуто в задачу %s", issueKey)
	contentJSON := map[string]interface{}{
		"type":     "doc",
		"children": []interface{}{mkParagraph(text)},
	}

	doc := &MessageDocument{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentJSON: contentJSON,
		ContentText: text,
		ThreadID:    threadID,
		CreatedAt:   time.Now().UTC(),
	}

	id, err := s.repo.InsertMessage(ctx, doc)
	if err != nil {
		return err
	}
	doc.ID = id

	// Инкремент reply_count родителя.
	if threadID != "" {
		if rcErr := s.repo.IncrementReplyCount(ctx, threadID); rcErr != nil {
			slog.Warn("promote-to-issue: failed to bump reply_count", "error", rcErr)
		}
	}

	// NATS-событие для real-time доставки.
	event := newMessageSentEvent(doc, nil)
	s.nats.PublishJSON(fmt.Sprintf(SubjectMessageSent, channelID.String()), event)
	s.publishUserNewMessageNotifications(ctx, doc)
	return nil
}
