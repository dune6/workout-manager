package server

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"workout-manager/internal/database"
	"workout-manager/internal/models/database/trainings"
)

type ITrainings interface {
	AddTraining(training trainings.Training) (primitive.ObjectID, error)
	DeleteTraining(id primitive.ObjectID) error
	GetUserTrainings(username string) ([]trainings.Training, error)
}

func (s *Server) AddTraining(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	const op = "server.AddTraining"

	var training trainings.Training
	err := json.NewDecoder(r.Body).Decode(&training)
	if err != nil {
		http.Error(w, "Ошибка при декодировании training"+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	id, err := s.DB.AddTraining(training)
	if err != nil {
		http.Error(w, op+": Ошибка при добавлении training "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Добавление успешно", "id": primitive.ObjectID.String(id)})
}

func (s *Server) GetUserTrainings(w http.ResponseWriter, r *http.Request, opt httprouter.Params) {
	const op = "server.GetUserTrainings"
	var resultTrainings []trainings.Training

	if opt.ByName("username") == "" {
		http.Error(w, op+": username required", http.StatusBadRequest)
		return
	}

	resultTrainings, err := s.DB.GetUserTrainings(opt.ByName("username"))
	if err != nil {
		http.Error(w, op+": ошибка получения тренировок: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resultTrainings) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Не найдены тренировки пользователя"})
		return
	}

	w.WriteHeader(http.StatusFound)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Получение успешно", "trainings": resultTrainings})
}

func (s *Server) DeleteTraining(w http.ResponseWriter, r *http.Request, opt httprouter.Params) {
	const op = "server.DeleteTraining"

	if opt.ByName("id") == "" {
		http.Error(w, op+": id required ", http.StatusBadRequest)
		return
	}

	index, err := primitive.ObjectIDFromHex(opt.ByName("id"))
	if err != nil {
		http.Error(w, op+": ошибка получения id: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = s.DB.DeleteTraining(index)
	if errors.Is(err, database.ErrorTrainingNotExist) {
		http.Error(w, op+": ошибка удаления тренировок: "+err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, op+": ошибка удаления тренировок: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Успешное удаление тренировки"})
}
