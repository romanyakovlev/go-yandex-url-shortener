package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/interceptors"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/middlewares"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	pb "github.com/romanyakovlev/go-yandex-url-shortener/internal/protobuf/protobuf"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/workers"
)

// URLShortener Интерфейс сервиса сокращения ссылок
type URLShortener interface {
	// AddURL добавление url
	AddURL(urlStr string) (models.SavedURL, error)
	// AddBatchURL добавление списка url
	AddBatchURL(batchArray []models.ShortenBatchURLRequestElement) ([]models.CorrelationSavedURL, error)
	// AddUserToURL присвоение url пользователю
	AddUserToURL(SavedURL models.SavedURL, user models.User) error
	// AddBatchUserToURL присвоение списка url пользователю
	AddBatchUserToURL(SavedURLs []models.SavedURL, user models.User) error
	// GetURL Получение url по короткой ссылке
	GetURL(shortURL string) (models.URLRow, bool)
	// GetURLByUser Получение всех url, присвоенных пользователю
	GetURLByUser(user models.User) ([]models.URLByUserResponseElement, bool)
	// GetURLByOriginalURL Получение короткой ссылки для url
	GetURLByOriginalURL(originalURL string) (string, bool)
	// DeleteBatchURL удаление списка url
	DeleteBatchURL(urls []string, user models.User) error
	// ConvertCorrelationSavedURLsToResponse преобразование модели данных []models.CorrelationSavedURL
	// в response-модель []models.ShortenBatchURLResponseElement для API-хелдлера
	ConvertCorrelationSavedURLsToResponse(correlationSavedURLs []models.CorrelationSavedURL) []models.ShortenBatchURLResponseElement
	// ConvertCorrelationSavedURLsToSavedURL преобразование модели данных []models.CorrelationSavedURL
	// в модель []models.SavedURL для API-хелдлера
	ConvertCorrelationSavedURLsToSavedURL(correlationSavedURLs []models.CorrelationSavedURL) []models.SavedURL
	// GetStats возвращает статистику
	GetStats() (models.URLStats, bool)
}

// Server представляет собой структуру grpc-сервера для обработки запросов
type Server struct {
	pb.UnimplementedURLShortenerServiceServer
	Shortener URLShortener
	Worker    *workers.URLDeletionWorker
}

// GetURLByID возвращает url на основе короткой ссылки
func (s *Server) GetURLByID(ctx context.Context, in *pb.GetURLByIDRequest) (*pb.GetURLByIDResponse, error) {
	urlRow, ok := s.Shortener.GetURL(in.ShortURL)
	if urlRow.DeletedFlag {
		return &pb.GetURLByIDResponse{}, status.Error(codes.NotFound, "URL is deleted")
	}
	if ok {
		return &pb.GetURLByIDResponse{OriginalURL: urlRow.OriginalURL}, nil
	} else {
		return &pb.GetURLByIDResponse{}, status.Error(codes.InvalidArgument, "Bad request")
	}
}

// SaveURL Принимает url и возвращает короткую ссылку (ожидает url в text/plain body)
func (s *Server) SaveURL(ctx context.Context, in *pb.SaveURLRequest) (*pb.SaveURLResponse, error) {
	savedURL, err := s.Shortener.AddURL(in.Url)
	if err != nil {
		//return s.handleShortenerServiceError(w, err, in.Url, "text")
		return &pb.SaveURLResponse{}, nil
	}
	user, _ := interceptors.GetUserFromContext(ctx)
	if err := s.Shortener.AddUserToURL(savedURL, user); err != nil {
		//s.handleError(w, err, http.StatusInternalServerError, "something went wrong: %s", nil)
		return &pb.SaveURLResponse{}, nil
	} else {
		return &pb.SaveURLResponse{ShortURL: savedURL.ShortURL}, nil
	}
}

// DeleteBatchURL Удаляет список url
func (s *Server) DeleteBatchURL(ctx context.Context, in *pb.DeleteBatchURLRequest) (*emptypb.Empty, error) {
	user, _ := middlewares.GetUserFromContext(ctx)
	req := workers.DeletionRequest{User: user, URLs: in.Urls}
	if err := s.Worker.SendDeletionRequestToWorker(req); err != nil {
		//s.handleError(w, err, http.StatusInternalServerError, "error sending to deletion worker request: %s", nil)
		return &emptypb.Empty{}, nil
	} else {
		return &emptypb.Empty{}, nil
	}
}

// GetURLByUser возвращает список url, которые пользователь загрузил в систему
func (s *Server) GetURLByUser(ctx context.Context, in *pb.GetURLByUserRequest) (*pb.GetURLByUserResponse, error) {
	user, _ := middlewares.GetUserFromContext(ctx)
	resp, ok := s.Shortener.GetURLByUser(user)
	if ok {
		return s.convertURLByUserResponseElements(resp), nil
	} else {
		return &pb.GetURLByUserResponse{}, status.Error(codes.InvalidArgument, "Bad request")
	}
}

// ShortenURL Принимает url и возвращает короткую ссылку (ожидает url в json body)
// дубликат SaveURL

// ShortenBatchURL Принимает список url в формате json и возвращает список коротких ссылок
func (s *Server) ShortenBatchURL(ctx context.Context, in *pb.ShortenBatchURLRequest) (*pb.ShortenBatchURLResponse, error) {
	user, _ := interceptors.GetUserFromContext(ctx)
	correlationSavedURLs, err := s.Shortener.AddBatchURL(s.convertPBShortenBatchURLRequest(in))
	resp := s.Shortener.ConvertCorrelationSavedURLsToResponse(correlationSavedURLs)
	if err != nil {
		//return s.handleError(w, err, http.StatusInternalServerError, "Shortener service error: %s", nil)
		return &pb.ShortenBatchURLResponse{}, nil
	}
	savedURLs := s.Shortener.ConvertCorrelationSavedURLsToSavedURL(correlationSavedURLs)
	if err := s.Shortener.AddBatchUserToURL(savedURLs, user); err != nil {
		//s.handleError(w, err, http.StatusInternalServerError, "something went wrong: %s", nil)
		return &pb.ShortenBatchURLResponse{}, nil
	}
	return s.convertShortenBatchURLResponseElement(resp), nil
}

// GetStats Возвращает статистику по количеству url-ов и пользователей
func (s *Server) GetStats(ctx context.Context, in *emptypb.Empty) (*pb.GetStatsResponse, error) {
	resp, ok := s.Shortener.GetStats()
	if ok {
		return &pb.GetStatsResponse{Urls: int32(resp.URLs), Users: int32(resp.Users)}, nil
	} else {
		return &pb.GetStatsResponse{}, status.Error(codes.InvalidArgument, "Bad request")
	}
}

/*
// handleError обарабатывает ошибки, возникающие при вызове методов в контроллере
func (s *Server) handleError(w http.ResponseWriter, err error, statusCode int, logMessage string, resp interface{}) {
	//c.logger.Debugf(logMessage, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if resp != nil {
		enc := json.NewEncoder(w)
		if encodeErr := enc.Encode(resp); encodeErr != nil {
			//c.logger.Debugf("cannot encode response JSON body: %s", encodeErr)
		}
	}
}

*/

/*
// handleShortenerServiceError обарабатывает специфичные ошибки URLShortener сервиса
func (s *Server) handleShortenerServiceError(w http.ResponseWriter, err error, urlStr string, message interface{}) (*pb.SaveURLResponse, error) {
	var appError *apperrors.OriginalURLAlreadyExists
	if ok := errors.As(err, &appError); ok {
		//logger.Debugf("Shortener service error: %s", err)
		value, ok := s.Shortener.GetURLByOriginalURL(urlStr)
		if !ok {
			return &message{}, status.Error(codes.InvalidArgument, "Bad request")
		}
		//resp := models.ShortenURLResponse{Result: value}
		//s.writeJSONResponse(w, http.StatusConflict, resp)
	} else {
		//c.logger.Debugf("Shortener service error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

*/

// convertURLByUserResponseElements конвертирует []models.URLByUserResponseElement
// в pb.GetURLByUserResponse protobuf объект
func (s *Server) convertURLByUserResponseElements(elements []models.URLByUserResponseElement) *pb.GetURLByUserResponse {
	response := &pb.GetURLByUserResponse{
		Urls: make([]*pb.URLByUserResponseElement, 0, len(elements)),
	}
	for _, elem := range elements {
		pbElem := &pb.URLByUserResponseElement{
			OriginalURL: elem.OriginalURL,
			ShortURL:    elem.ShortURL,
		}
		response.Urls = append(response.Urls, pbElem)
	}
	return response
}

// ConvertPBShortenBatchURLRequest конвертирует pb.ShortenBatchURLRequest
// в []models.ShortenBatchURLRequestElement.
func (s *Server) convertPBShortenBatchURLRequest(pbRequest *pb.ShortenBatchURLRequest) []models.ShortenBatchURLRequestElement {
	var modelElements []models.ShortenBatchURLRequestElement

	for _, pbElement := range pbRequest.Urls {
		modelElement := models.ShortenBatchURLRequestElement{
			OriginalURL:   pbElement.OriginalURL,
			CorrelationID: pbElement.CorrelationId,
		}

		modelElements = append(modelElements, modelElement)
	}

	return modelElements
}

// convertShortenBatchURLResponseElement конвертирует []models.ShortenBatchURLResponseElement
// в pb.ShortenBatchURLResponse protobuf объект.
func (s *Server) convertShortenBatchURLResponseElement(modelElements []models.ShortenBatchURLResponseElement) *pb.ShortenBatchURLResponse {
	pbResponse := &pb.ShortenBatchURLResponse{
		Urls: make([]*pb.ShortenBatchURLResponseElement, 0, len(modelElements)),
	}

	for _, modelElement := range modelElements {
		pbElement := &pb.ShortenBatchURLResponseElement{
			CorrelationId: modelElement.CorrelationID,
			ShortURL:      modelElement.ShortURL,
		}

		pbResponse.Urls = append(pbResponse.Urls, pbElement)
	}

	return pbResponse
}
