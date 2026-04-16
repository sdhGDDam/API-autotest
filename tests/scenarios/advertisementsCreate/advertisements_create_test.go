package tests

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	advHelper "api-tests-template/internal/helpers/advertisements_helper"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type AdvertisementsCreateTestSuite struct {
	suite.Suite
	token      string
	userID     string
	createdIDs []string
	photoPaths []string
}

func TestAdvertisementsCreateSuite(t *testing.T) {
	suite.Run(t, &AdvertisementsCreateTestSuite{})
}

func (s *AdvertisementsCreateTestSuite) SetupSuite() {
	base.SetupSuite()
	base.Precondition("Авторизация пользователя")
	loginData := auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.token = loginData.Token
	s.userID = loginData.User.Id

	dir, err := os.Getwd()
	require.NoError(s.T(), err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			s.T().Fatal("go.mod not found")
		}
		dir = parent
	}
	s.photoPaths = []string{
		filepath.Join(dir, "testdata", "photo1.jpg"),
		filepath.Join(dir, "testdata", "photo2.png"),
		filepath.Join(dir, "testdata", "photo3.webp"),
	}
	for _, p := range s.photoPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			s.T().Fatalf("Тестовое фото не найдено: %s", p)
		}
	}
	s.createdIDs = make([]string, 0)
}

func (s *AdvertisementsCreateTestSuite) TearDownSuite() {
	base.Precondition("Удаление всех созданных объявлений")
	for _, id := range s.createdIDs {
		advHelper.DeleteAdvertisement(s.T(), s.token, id)
	}
	base.TearDownSuite()
}

//
// ПОЗИТИВНЫЕ ТЕСТЫ
//

// Позитивный: создание объявления со всеми параметрами
func (s *AdvertisementsCreateTestSuite) TestPositiveCreateAdvertisement() {
	title := "Создание " + utils.RandomString(5)
	desc := "Описание товара"
	price := "999.99"
	qty := "10"

	fields := map[string]string{
		"title": title, "description": desc, "price": price, "quantity": qty,
	}
	resp := advHelper.CreateAdvertisement(s.T(), s.token, fields, s.photoPaths)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		s.T().Fatalf("Ожидался статус 201 Created, получен %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	adID := gjson.GetBytes(body, "id").String()
	if adID == "" {
		s.T().Fatal("ID созданного объявления пуст")
	}
	s.createdIDs = append(s.createdIDs, adID)

	if gjson.GetBytes(body, "title").String() != title {
		s.T().Fatalf("title: ожидался %s, получен %s", title, gjson.GetBytes(body, "title").String())
	}
	if gjson.GetBytes(body, "description").String() != desc {
		s.T().Fatalf("description: ожидалось %s, получено %s", desc, gjson.GetBytes(body, "description").String())
	}
	if gjson.GetBytes(body, "price").String() != price {
		s.T().Fatalf("price: ожидался %s, получен %s", price, gjson.GetBytes(body, "price").String())
	}
	if gjson.GetBytes(body, "quantity").String() != qty {
		s.T().Fatalf("quantity: ожидалось %s, получено %s", qty, gjson.GetBytes(body, "quantity").String())
	}
	if gjson.GetBytes(body, "user_id").String() != s.userID {
		s.T().Fatalf("user_id: ожидался %s, получен %s", s.userID, gjson.GetBytes(body, "user_id").String())
	}
	if len(gjson.GetBytes(body, "photos").Array()) != 3 {
		s.T().Fatalf("photos: ожидалось 3 фото, получено %d", len(gjson.GetBytes(body, "photos").Array()))
	}
}

// Позитивный: получение объявления по ID
func (s *AdvertisementsCreateTestSuite) TestPositiveGetAdvertisementById() {
	title := "Для получения " + utils.RandomString(5)
	fields := map[string]string{"title": title, "description": "desc"}
	respCreate := advHelper.CreateAdvertisement(s.T(), s.token, fields, s.photoPaths)
	defer respCreate.Body.Close()
	if respCreate.StatusCode != http.StatusCreated {
		s.T().Fatalf("Не удалось создать объявление для теста: статус %d", respCreate.StatusCode)
	}
	bodyCreate, _ := io.ReadAll(respCreate.Body)
	adID := gjson.GetBytes(bodyCreate, "id").String()
	s.createdIDs = append(s.createdIDs, adID)

	respGet := advHelper.GetAdvertisementById(s.T(), adID)
	defer respGet.Body.Close()
	if respGet.StatusCode != http.StatusOK {
		s.T().Fatalf("GET /advertisement?id=%s вернул статус %d, ожидался 200", adID, respGet.StatusCode)
	}
	bodyGet, _ := io.ReadAll(respGet.Body)
	if gjson.GetBytes(bodyGet, "id").String() != adID {
		s.T().Fatalf("ID в ответе (%s) не совпадает с запрошенным (%s)", gjson.GetBytes(bodyGet, "id").String(), adID)
	}
	if gjson.GetBytes(bodyGet, "title").String() != title {
		s.T().Fatalf("title в ответе (%s) не совпадает с ожидаемым (%s)", gjson.GetBytes(bodyGet, "title").String(), title)
	}
}

// Позитивный: получение фото через GET /advertisements/{id}/photos
func (s *AdvertisementsCreateTestSuite) TestPositiveGetAdvertisementPhotos() {
	title := "Для фото " + utils.RandomString(5)
	fields := map[string]string{"title": title, "description": "desc"}
	respCreate := advHelper.CreateAdvertisement(s.T(), s.token, fields, s.photoPaths)
	defer respCreate.Body.Close()
	if respCreate.StatusCode != http.StatusCreated {
		s.T().Fatalf("Не удалось создать объявление для теста фото: статус %d", respCreate.StatusCode)
	}
	bodyCreate, _ := io.ReadAll(respCreate.Body)
	adID := gjson.GetBytes(bodyCreate, "id").String()
	s.createdIDs = append(s.createdIDs, adID)

	respPhotos := advHelper.GetAdvertisementPhotos(s.T(), s.token, adID)
	defer respPhotos.Body.Close()
	if respPhotos.StatusCode != http.StatusOK {
		s.T().Fatalf("GET /advertisements/%s/photos вернул статус %d, ожидался 200", adID, respPhotos.StatusCode)
	}
	bodyPhotos, _ := io.ReadAll(respPhotos.Body)
	photos := gjson.ParseBytes(bodyPhotos).Array()
	if len(photos) != 3 {
		s.T().Fatalf("Ожидалось 3 фото, получено %d", len(photos))
	}
	for i, ph := range photos {
		if ph.Get("url").String() == "" {
			s.T().Fatalf("Фото %d не имеет URL", i)
		}
	}
}

// Позитивный: поиск объявления по полному названию
func (s *AdvertisementsCreateTestSuite) TestPositiveSearchAdvertisement() {
	title := "Для поиска " + utils.RandomString(5)
	fields := map[string]string{"title": title, "description": "desc"}
	respCreate := advHelper.CreateAdvertisement(s.T(), s.token, fields, s.photoPaths)
	defer respCreate.Body.Close()
	if respCreate.StatusCode != http.StatusCreated {
		s.T().Fatalf("Не удалось создать объявление для теста поиска: статус %d", respCreate.StatusCode)
	}
	bodyCreate, _ := io.ReadAll(respCreate.Body)
	adID := gjson.GetBytes(bodyCreate, "id").String()
	s.createdIDs = append(s.createdIDs, adID)

	respSearch := advHelper.SearchAdvertisements(s.T(), title)
	defer respSearch.Body.Close()
	if respSearch.StatusCode != http.StatusOK {
		s.T().Fatalf("GET /advertisements?search=%s вернул статус %d, ожидался 200", title, respSearch.StatusCode)
	}
	bodySearch, _ := io.ReadAll(respSearch.Body)
	items := gjson.GetBytes(bodySearch, "items").Array()
	found := false
	for _, it := range items {
		if it.Get("id").String() == adID {
			found = true
			break
		}
	}
	if !found {
		s.T().Fatalf("Объявление с ID %s не найдено в поиске по названию '%s'", adID, title)
	}
}

//
// НЕГАТИВНЫЕ ТЕСТЫ
//

// Негативный: попытка создания объявления без токена авторизации
func (s *AdvertisementsCreateTestSuite) TestNegativeCreateAdvertisementUnauthorized() {
	fields := map[string]string{"title": "x", "description": "x"}
	resp := advHelper.CreateAdvertisement(s.T(), "", fields, s.photoPaths)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		s.T().Fatalf("Ожидался статус 401 или 403, получен %d", resp.StatusCode)
	}
}

// Негативный: отсутствие обязательных полей
func (s *AdvertisementsCreateTestSuite) TestNegativeMissingRequiredFields() {
	cases := []struct {
		name   string
		fields map[string]string
		expect string
	}{
		{"Без title", map[string]string{"description": "desc"}, "title"},
		{"Без description", map[string]string{"title": "title"}, "description"},
		{"Без photos", map[string]string{"title": "t", "description": "d"}, "photo"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			var photos []string
			if tc.expect == "photo" {
				photos = []string{}
			} else {
				photos = s.photoPaths
			}
			resp := advHelper.CreateAdvertisement(s.T(), s.token, tc.fields, photos)
			defer resp.Body.Close()
			if resp.StatusCode < 400 {
				s.T().Fatalf("Ожидался статус ошибки (4xx/5xx), получен %d", resp.StatusCode)
			}
			body, _ := io.ReadAll(resp.Body)
			msg := gjson.GetBytes(body, "message").String()
			if msg == "" {
				s.T().Fatalf("Тело ответа не содержит message: %s", string(body))
			}
			if !utils.ContainsSubstring(msg, tc.expect) {
				s.T().Fatalf("Сообщение '%s' не содержит ожидаемую подстроку '%s'", msg, tc.expect)
			}
		})
	}
}

// Негативный: некорректные значения
func (s *AdvertisementsCreateTestSuite) TestNegativeInvalidValues() {
	valid := map[string]string{
		"title": "ok", "description": "ok", "price": "100", "quantity": "5",
	}
	s.Run("Отрицательная цена", func() {
		f := utils.CopyMap(valid)
		f["price"] = "-10"
		resp := advHelper.CreateAdvertisement(s.T(), s.token, f, s.photoPaths)
		defer resp.Body.Close()
		if resp.StatusCode < 400 {
			s.T().Fatalf("Ожидался статус ошибки, получен %d", resp.StatusCode)
		}
	})
	s.Run("Нулевое количество", func() {
		f := utils.CopyMap(valid)
		f["quantity"] = "0"
		resp := advHelper.CreateAdvertisement(s.T(), s.token, f, s.photoPaths)
		defer resp.Body.Close()
		if resp.StatusCode < 400 {
			s.T().Fatalf("Ожидался статус ошибки, получен %d", resp.StatusCode)
		}
	})
	s.Run("Слишком длинный title (>255 символов)", func() {
		f := utils.CopyMap(valid)
		long := make([]byte, 300)
		for i := range long {
			long[i] = 'A'
		}
		f["title"] = string(long)
		resp := advHelper.CreateAdvertisement(s.T(), s.token, f, s.photoPaths)
		defer resp.Body.Close()
		if resp.StatusCode < 400 {
			s.T().Fatalf("Ожидался статус ошибки, получен %d", resp.StatusCode)
		}
	})
	s.Run("4 фото (превышение лимита)", func() {
		four := append(s.photoPaths, s.photoPaths[0])
		resp := advHelper.CreateAdvertisement(s.T(), s.token, valid, four)
		defer resp.Body.Close()
		if resp.StatusCode < 400 {
			s.T().Fatalf("Ожидался статус ошибки, получен %d", resp.StatusCode)
		}
	})
}
