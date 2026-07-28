package crm

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/crmcontact"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — доступ к данным CRM.
type Repository interface {
	CreateContact(ctx context.Context, userID uuid.UUID, req CreateContactRequest) (*ent.CrmContact, error)
	// ListContacts — ownerID == nil → org-wide (admin); иначе только created_by = ownerID (RT-003).
	ListContacts(ctx context.Context, ownerID *uuid.UUID) ([]*ent.CrmContact, error)
	GetContact(ctx context.Context, id uuid.UUID) (*ent.CrmContact, error)
	UpdateContact(ctx context.Context, id uuid.UUID, req UpdateContactRequest) (*ent.CrmContact, error)
	SoftDeleteContact(ctx context.Context, id uuid.UUID) error

	CreateDeal(ctx context.Context, ownerID uuid.UUID, req CreateDealRequest) (*ent.CrmDeal, error)
	// ListDeals — ownerID == nil → org-wide (admin); иначе только owner = ownerID (RT-003).
	ListDeals(ctx context.Context, ownerID *uuid.UUID) ([]*ent.CrmDeal, error)
	GetDeal(ctx context.Context, id uuid.UUID) (*ent.CrmDeal, error)
	UpdateDeal(ctx context.Context, id uuid.UUID, req UpdateDealRequest) (*ent.CrmDeal, error)
	SoftDeleteDeal(ctx context.Context, id uuid.UUID) error

	IsProjectMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

// EntRepository реализует Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий CRM.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateContact(ctx context.Context, userID uuid.UUID, req CreateContactRequest) (*ent.CrmContact, error) {
	create := r.client.CrmContact.Create().
		SetName(req.Name).
		SetCreatedByID(userID)
	if req.Email != "" {
		create.SetEmail(req.Email)
	}
	if req.Phone != "" {
		create.SetPhone(req.Phone)
	}
	if req.Company != "" {
		create.SetCompany(req.Company)
	}
	if req.Notes != "" {
		create.SetNotes(req.Notes)
	}
	if req.LinkedUserID != nil {
		create.SetLinkedUserID(*req.LinkedUserID)
	}
	return create.Save(ctx)
}

// MaxCRMList — P3-BE-06 / W7-A2-013: hard cap ListContacts / ListDeals (как accounting invoices 50).
// Mutation: снять Limit → unbounded All при N>cap.
const MaxCRMList = 100

func (r *EntRepository) ListContacts(ctx context.Context, ownerID *uuid.UUID) ([]*ent.CrmContact, error) {
	q := r.client.CrmContact.Query().
		Where(crmcontact.DeletedAtIsNil()).
		WithLinkedUser().
		Order(ent.Desc(crmcontact.FieldUpdatedAt)).
		Limit(MaxCRMList)
	if ownerID != nil {
		q = q.Where(crmcontact.HasCreatedByWith(user.IDEQ(*ownerID)))
	}
	return q.All(ctx)
}

func (r *EntRepository) GetContact(ctx context.Context, id uuid.UUID) (*ent.CrmContact, error) {
	return r.client.CrmContact.Query().
		Where(crmcontact.ID(id), crmcontact.DeletedAtIsNil()).
		WithLinkedUser().
		WithCreatedBy().
		Only(ctx)
}

func (r *EntRepository) UpdateContact(ctx context.Context, id uuid.UUID, req UpdateContactRequest) (*ent.CrmContact, error) {
	// A6-003: VersionEQ в WHERE — иначе concurrent update = lost write (app-check TOCTOU).
	upd := r.client.CrmContact.UpdateOneID(id).
		Where(
			crmcontact.DeletedAtIsNil(),
			crmcontact.VersionEQ(req.Version),
		).
		AddVersion(1)
	if req.Name != nil {
		upd.SetName(*req.Name)
	}
	if req.Email != nil {
		upd.SetEmail(*req.Email)
	}
	if req.Phone != nil {
		upd.SetPhone(*req.Phone)
	}
	if req.Company != nil {
		upd.SetCompany(*req.Company)
	}
	if req.Notes != nil {
		upd.SetNotes(*req.Notes)
	}
	if req.LinkedUserID != nil {
		upd.SetLinkedUserID(*req.LinkedUserID)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteContact(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.CrmContact.UpdateOneID(id).
		Where(crmcontact.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

func (r *EntRepository) CreateDeal(ctx context.Context, ownerID uuid.UUID, req CreateDealRequest) (*ent.CrmDeal, error) {
	create := r.client.CrmDeal.Create().
		SetTitle(req.Title).
		SetOwnerID(ownerID)
	if req.Description != "" {
		create.SetDescription(req.Description)
	}
	if req.Stage != "" {
		create.SetStage(crmdeal.Stage(req.Stage))
	}
	if req.AmountCents != nil {
		create.SetAmountCents(*req.AmountCents)
	}
	if req.Currency != "" {
		create.SetCurrency(req.Currency)
	}
	if req.ContactID != nil {
		create.SetContactID(*req.ContactID)
	}
	if req.ProjectID != nil {
		create.SetProjectID(*req.ProjectID)
	}
	return create.Save(ctx)
}

func (r *EntRepository) ListDeals(ctx context.Context, ownerID *uuid.UUID) ([]*ent.CrmDeal, error) {
	q := r.client.CrmDeal.Query().
		Where(crmdeal.DeletedAtIsNil()).
		WithOwner().
		WithContact().
		WithProject().
		Order(
			ent.Asc(crmdeal.FieldStage),
			ent.Asc(crmdeal.FieldRank),
			ent.Desc(crmdeal.FieldUpdatedAt),
		).
		Limit(MaxCRMList)
	if ownerID != nil {
		q = q.Where(crmdeal.HasOwnerWith(user.IDEQ(*ownerID)))
	}
	return q.All(ctx)
}

func (r *EntRepository) GetDeal(ctx context.Context, id uuid.UUID) (*ent.CrmDeal, error) {
	return r.client.CrmDeal.Query().
		Where(crmdeal.ID(id), crmdeal.DeletedAtIsNil()).
		WithOwner().
		WithContact().
		WithProject().
		Only(ctx)
}

func (r *EntRepository) UpdateDeal(ctx context.Context, id uuid.UUID, req UpdateDealRequest) (*ent.CrmDeal, error) {
	// A6-003: VersionEQ в WHERE — иначе concurrent update = lost write (app-check TOCTOU).
	upd := r.client.CrmDeal.UpdateOneID(id).
		Where(
			crmdeal.DeletedAtIsNil(),
			crmdeal.VersionEQ(req.Version),
		).
		AddVersion(1)
	if req.Title != nil {
		upd.SetTitle(*req.Title)
	}
	if req.Description != nil {
		upd.SetDescription(*req.Description)
	}
	if req.Stage != nil {
		upd.SetStage(crmdeal.Stage(*req.Stage))
	}
	if req.AmountCents != nil {
		upd.SetAmountCents(*req.AmountCents)
	}
	if req.Currency != nil {
		upd.SetCurrency(*req.Currency)
	}
	if req.ContactID != nil {
		upd.SetContactID(*req.ContactID)
	}
	if req.ProjectID != nil {
		upd.SetProjectID(*req.ProjectID)
	}
	if req.Rank != nil {
		upd.SetRank(*req.Rank)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteDeal(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.CrmDeal.UpdateOneID(id).
		Where(crmdeal.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

func (r *EntRepository) IsProjectMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	return r.client.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
}

func (r *EntRepository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return r.client.User.Query().
		Where(user.IDEQ(userID)).
		Exist(ctx)
}

func (r *EntRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleAdmin || u.SystemRole == user.SystemRoleFounder, nil
}
