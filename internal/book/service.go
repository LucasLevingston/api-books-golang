package book

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(filter ListFilter) (PageResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}
	allowed := map[string]bool{"name": true, "author": true, "masterpiece": true, "year": true}
	if !allowed[filter.SortBy] {
		filter.SortBy = "name"
	}
	if filter.SortDir != "asc" && filter.SortDir != "desc" {
		filter.SortDir = "asc"
	}
	return s.repo.FindAll(filter)
}

func (s *service) Get(id int64) (Book, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(input CreateInput) (Book, error) {
	return s.repo.Create(input)
}

func (s *service) Update(id int64, input UpdateInput) (Book, error) {
	return s.repo.Update(id, input)
}

func (s *service) Delete(id int64) error {
	return s.repo.Delete(id)
}
