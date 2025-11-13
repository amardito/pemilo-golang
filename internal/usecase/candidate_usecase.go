package usecase

import (
	"time"

	"github.com/amardito/pemilo-golang/internal/domain"
	"github.com/google/uuid"
)

type CandidateUsecase struct {
	candidateRepo    domain.CandidateRepository
	subCandidateRepo domain.SubCandidateRepository
	roomRepo         domain.RoomRepository
}

func NewCandidateUsecase(candidateRepo domain.CandidateRepository, subCandidateRepo domain.SubCandidateRepository, roomRepo domain.RoomRepository) *CandidateUsecase {
	return &CandidateUsecase{
		candidateRepo:    candidateRepo,
		subCandidateRepo: subCandidateRepo,
		roomRepo:         roomRepo,
	}
}

func (u *CandidateUsecase) CreateCandidate(roomID, name, photoURL, description string, subCandidates []struct {
	Name        string
	PhotoURL    string
	Description string
}) (*domain.CandidateWithSubs, error) {
	// Verify room exists
	_, err := u.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, err
	}

	// Create candidate
	candidate := &domain.Candidate{
		ID:          uuid.New().String(),
		RoomID:      roomID,
		Name:        name,
		PhotoURL:    photoURL,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.candidateRepo.Create(candidate); err != nil {
		return nil, err
	}

	// Create sub-candidates
	var subs []domain.SubCandidate
	for _, sub := range subCandidates {
		subCandidate := &domain.SubCandidate{
			ID:          uuid.New().String(),
			CandidateID: candidate.ID,
			Name:        sub.Name,
			PhotoURL:    sub.PhotoURL,
			Description: sub.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := u.subCandidateRepo.Create(subCandidate); err != nil {
			return nil, err
		}
		subs = append(subs, *subCandidate)
	}

	return &domain.CandidateWithSubs{
		Candidate:     *candidate,
		SubCandidates: subs,
	}, nil
}

func (u *CandidateUsecase) GetCandidate(id string) (*domain.CandidateWithSubs, error) {
	candidate, err := u.candidateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	subs, err := u.subCandidateRepo.GetByCandidateID(id)
	if err != nil {
		return nil, err
	}

	subsList := make([]domain.SubCandidate, 0, len(subs))
	for _, sub := range subs {
		subsList = append(subsList, *sub)
	}

	return &domain.CandidateWithSubs{
		Candidate:     *candidate,
		SubCandidates: subsList,
	}, nil
}

func (u *CandidateUsecase) GetCandidatesByRoom(roomID string) ([]*domain.CandidateWithSubs, error) {
	candidates, err := u.candidateRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, err
	}

	var result []*domain.CandidateWithSubs
	for _, candidate := range candidates {
		subs, err := u.subCandidateRepo.GetByCandidateID(candidate.ID)
		if err != nil {
			return nil, err
		}

		subsList := make([]domain.SubCandidate, 0, len(subs))
		for _, sub := range subs {
			subsList = append(subsList, *sub)
		}

		result = append(result, &domain.CandidateWithSubs{
			Candidate:     *candidate,
			SubCandidates: subsList,
		})
	}

	return result, nil
}

func (u *CandidateUsecase) UpdateCandidate(id string, name, photoURL, description *string) (*domain.Candidate, error) {
	candidate, err := u.candidateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		candidate.Name = *name
	}
	if photoURL != nil {
		candidate.PhotoURL = *photoURL
	}
	if description != nil {
		candidate.Description = *description
	}

	if err := u.candidateRepo.Update(candidate); err != nil {
		return nil, err
	}

	return candidate, nil
}

func (u *CandidateUsecase) DeleteCandidate(id string) error {
	return u.candidateRepo.Delete(id)
}

func (u *CandidateUsecase) UpdateSubCandidate(id string, name, photoURL, description *string) (*domain.SubCandidate, error) {
	subCandidate, err := u.subCandidateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		subCandidate.Name = *name
	}
	if photoURL != nil {
		subCandidate.PhotoURL = *photoURL
	}
	if description != nil {
		subCandidate.Description = *description
	}

	if err := u.subCandidateRepo.Update(subCandidate); err != nil {
		return nil, err
	}

	return subCandidate, nil
}

func (u *CandidateUsecase) DeleteSubCandidate(id string) error {
	return u.subCandidateRepo.Delete(id)
}
