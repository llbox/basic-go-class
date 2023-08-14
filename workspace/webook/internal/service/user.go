package service

import (
	"basic-go-class/workspace/webook/internal/domain"
	"basic-go-class/workspace/webook/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("无效的账号或密码")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) SignIn(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindUserByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	log.Printf("user: %+v", u)
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *UserService) Edit(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateUserById(ctx, u)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindUserById(ctx, id)
}
