package server

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"workout-manager/internal/database"
	models "workout-manager/internal/models/database/user"
)

type IRegister interface {
	Register(user models.User) error
}

type ILogin interface {
	Login(username, password string) (*models.User, error)
}

// Register Регистрация пользователя
func (s *Server) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Ошибка при хэшировании пароля", http.StatusInternalServerError)
		return
	}

	user.Password = hashedPassword
	user.ID = primitive.NewObjectID()

	// Сохраняем пользователя в базу
	err = s.DB.Register(user)
	if err != nil {
		switch {
		//case errors.Is(err, database.ErrorUserNotFound):
		//	http.Error(w, "Пользователь уже существует", http.StatusConflict)
		//	return
		case errors.Is(err, database.ErrorUserExist):
			http.Error(w, "Пользователь уже существует", http.StatusConflict)
			return
		case errors.Is(err, database.ErrorUserInsert):
			http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Регистрация успешна"})
}

// Login Логин пользователя
func (s *Server) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input models.User

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	findUser, err := s.DB.Login(input.Username, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrorUserNotFound):
			http.Error(w, "Неверный логин или пароль", http.StatusConflict)
			return
		case errors.Is(err, database.ErrorSomethingGetWrong):
			http.Error(w, "Ошибка при обработке запроса", http.StatusInternalServerError)
			return
		}
	}

	// Проверяем пароль
	if !CheckPassword(findUser.Password, input.Password) {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Авторизация успешна"})
}

// HashPassword Хэширование пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword Проверка пароля
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
