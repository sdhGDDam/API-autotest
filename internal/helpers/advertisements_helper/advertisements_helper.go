package advertisements_helper

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"

	"github.com/stretchr/testify/require"
)

// CreateAdvertisement отправляет POST /advertisement с multipart/form-data.
// Параметры:
//   - t: *testing.T для ассертов
//   - token: JWT токен (пустая строка – без авторизации)
//   - fields: map полей формы (title, description, price, quantity)
//   - photoPaths: срез путей к файлам для поля photos
// Возвращает сырой HTTP-ответ.
func CreateAdvertisement(t *testing.T, token string, fields map[string]string, photoPaths []string) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range fields {
		_ = writer.WriteField(k, v)
	}
	for _, p := range photoPaths {
		file, err := os.Open(p)
		require.NoError(t, err)
		defer file.Close()
		part, err := writer.CreateFormFile("photos", filepath.Base(p))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	return apiRunner.GetRunner().Auth(token).Create().
		Post(path.AdvertisementPath).
		ContentType(writer.FormDataContentType()).
		Body(body.String()).
		Expect(t).
		End().Response
}

// GetAdvertisementById выполняет GET /advertisement?id=...
func GetAdvertisementById(t *testing.T, id string) *http.Response {
	return apiRunner.GetRunner().Create().
		Get(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}

// SearchAdvertisements выполняет GET /advertisements?search=...
func SearchAdvertisements(t *testing.T, search string) *http.Response {
	return apiRunner.GetRunner().Create().
		Get(path.AdvertisementsPath).
		Query("search", search).
		Expect(t).
		End().Response
}

// DeleteAdvertisement выполняет DELETE /advertisement?id=... и проверяет успешность.
func DeleteAdvertisement(t *testing.T, token, id string) {
	resp := apiRunner.GetRunner().Auth(token).Create().
		Delete(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
	defer resp.Body.Close()
	require.Contains(t, []int{http.StatusNoContent, http.StatusOK}, resp.StatusCode)
}

// GetAdvertisementPhotos выполняет GET /advertisements/{id}/photos с авторизацией.
func GetAdvertisementPhotos(t *testing.T, token, adID string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().
		Get("/advertisements/" + adID + "/photos").
		Expect(t).
		End().Response
}