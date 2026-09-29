package inmemory_test

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ya_url_shortener/internal/model"
	filestorage "ya_url_shortener/internal/repository/file_storage"
	inmemory "ya_url_shortener/internal/repository/in_memory"
)

type batch struct {
	input  []model.Resource
	output []model.ResourceBatchOutput
}

func newTestStore(t *testing.T) (*inmemory.Store, *filestorage.FileStorage, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "storage.txt")
	fStorage, err := filestorage.NewFileStorage(path)
	require.NoError(t, err)
	s, err := inmemory.NewStore(fStorage)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fStorage.Close() })
	return s, fStorage, path
}

func TestCreateResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want model.Resource
	}{
		{
			name: "happy CreateResource #1",
			want: model.Resource{
				ID:          1,
				OriginalURL: "http://ya.ru",
				ShortURL:    "mocked",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s, fStorage, path := newTestStore(t)
			got, err := s.CreateResource(t.Context(), test.want)
			require.NoError(t, err)
			require.Equal(t, test.want.OriginalURL, got.OriginalURL)
			require.Equal(t, test.want.ShortURL, got.ShortURL)
			require.NoError(t, fStorage.Close()) // освобождаем файл перед повторным открытием

			// переоткрываем файл и пересоздаем репозиторий: должен сработать бэкфилл
			fStorage, err = filestorage.NewFileStorage(path)
			require.NoError(t, err)
			recreatedStore, err := inmemory.NewStore(fStorage)
			require.NoError(t, err)
			t.Cleanup(func() { _ = fStorage.Close() })

			got, err = recreatedStore.GetResourceByID(t.Context(), test.want.ID)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGetResourceByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    model.Resource
		wantErr error
	}{
		{
			name: "happy GetResourceByID #1",
			want: model.Resource{
				ID:          1,
				OriginalURL: "http://ya.ru",
				ShortURL:    "mocked",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s, _, _ := newTestStore(t)
			created, err := s.CreateResource(t.Context(), test.want)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			got, err := s.GetResourceByID(t.Context(), test.want.ID)
			require.NoError(t, err)
			assert.Equal(t, created, got)
		})
	}
}

func TestGetResourceByURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    model.Resource
		wantErr error
	}{
		{
			name: "happy GetResourceByURL #1",
			want: model.Resource{
				ID:          1,
				OriginalURL: "http://ya.ru",
				ShortURL:    "mocked",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s, _, _ := newTestStore(t)
			created, err := s.CreateResource(t.Context(), test.want)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			got, err := s.GetResourceByURL(t.Context(), created.ShortURL)
			require.NoError(t, err)
			assert.Equal(t, created, got)
		})
	}
}

func TestCreateBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want batch
	}{
		{
			name: "happy CreateBatch #1",
			want: batch{
				input: []model.Resource{
					{
						ID:            1,
						OriginalURL:   "http://ya.ru",
						ShortURL:      "mocked",
						CorrelationID: uuid.MustParse("0cc3dd37-e05f-49ab-a716-e783556a7980"),
					},
					{
						ID:            2,
						OriginalURL:   "http://yandex.ru",
						ShortURL:      "mocked2",
						CorrelationID: uuid.MustParse("241af749-05ea-42d9-8ada-3786c6d6018b"),
					},
				},
				output: []model.ResourceBatchOutput{
					{CorrelationID: uuid.MustParse("0cc3dd37-e05f-49ab-a716-e783556a7980"), ShortURL: "mocked"},
					{CorrelationID: uuid.MustParse("241af749-05ea-42d9-8ada-3786c6d6018b"), ShortURL: "mocked2"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s, fStorage, path := newTestStore(t)
			got, err := s.CreateBatch(t.Context(), test.want.input)
			require.NoError(t, err)
			require.ElementsMatch(t, test.want.output, got)
			require.NoError(t, fStorage.Close()) // освобождаем файл перед повторным открытием

			// переоткрываем файл и пересоздаем репозиторий: должен сработать бэкфилл
			fStorage, err = filestorage.NewFileStorage(path)
			require.NoError(t, err)
			recreatedStore, err := inmemory.NewStore(fStorage)
			require.NoError(t, err)
			t.Cleanup(func() { _ = fStorage.Close() })

			refilled, err := recreatedStore.GetResourceByID(t.Context(), test.want.input[0].ID)
			require.NoError(t, err)
			assert.Equal(t, test.want.input[0], refilled)
		})
	}
}
