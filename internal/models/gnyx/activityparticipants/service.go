package activityparticipants

type Service[T any] interface {
	GetAll() ([]T, error)
	GetByID(id string) (*ActivityParticipants, error)
	Create(activityParticipant *ActivityParticipants) error
	Update(activityParticipant *ActivityParticipants) error
	Delete(id string) error
}

type ActivityParticipantsService[T ActivityParticipants] struct {
	repo ORMRepository[T]
}

func NewService[T ActivityParticipants](repo ORMRepository[T]) Service[T] {
	return &ActivityParticipantsService[T]{repo: repo}
}

func (s *ActivityParticipantsService[T]) GetAll() ([]T, error) {
	return s.repo.GetAll()
}

func (s *ActivityParticipantsService[T]) GetByID(id string) (*ActivityParticipants, error) {
	return s.repo.GetByID(id)
}

func (s *ActivityParticipantsService[T]) Create(activityParticipant *ActivityParticipants) error {
	return s.repo.Create(activityParticipant)
}

func (s *ActivityParticipantsService[T]) Update(activityParticipant *ActivityParticipants) error {
	return s.repo.Update(activityParticipant)
}

func (s *ActivityParticipantsService[T]) Delete(id string) error {
	return s.repo.Delete(id)
}
