package service

func (s *Service) Favorite(slug string, userID uint, add bool) (ArticleView, error) {
	a, err := s.Repo.ArticleBySlug(slug)
	if err != nil {
		return ArticleView{}, storageError(err)
	}
	if add {
		err = s.Repo.AddFavorite(a.ID, userID)
	} else {
		err = s.Repo.RemoveFavorite(a.ID, userID)
	}
	if err != nil {
		return ArticleView{}, storageError(err)
	}
	return s.Article(slug, userID)
}
