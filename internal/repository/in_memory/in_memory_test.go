package inmemory_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ya_url_shortener/internal/model"
	filestorage "ya_url_shortener/internal/repository/file_storage"
	inmemory "ya_url_shortener/internal/repository/in_memory"
)

func newTestStore(t *testing.T) (*inmemory.Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "storage.txt")
	fStorage, err := filestorage.NewFileStorage(path)
	require.NoError(t, err)
	s, err := inmemory.NewStore(fStorage)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			_ = s.Close()
		})
	return s, path
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
				ID:        1,
				Address:   "http://ya.ru",
				Shortened: "mocked",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s, path := newTestStore(t)
			got, err := s.CreateResource(t.Context(), test.want)
			require.NoError(t, err)
			require.Equal(t, test.want.Address, got.Address)
			require.Equal(t, test.want.Shortened, got.Shortened)
			require.NoError(t, s.Close()) // закрываем репозиторий

			// переоткрываем файл и пересоздаем репозиторий: должен сработать бэкфилл
			// этот блок за пределами видимости хелпера
			fStorage, err := filestorage.NewFileStorage(path)
			require.NoError(t, err)
			recreatedStore, err := inmemory.NewStore(fStorage)
			require.NoError(t, err)
			// принудительный клинап пересозданного стора
			t.Cleanup(func() { recreatedStore.Close() })

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
				ID:        1,
				Address:   "http://ya.ru",
				Shortened: "mocked",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s, _ := newTestStore(t)
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
				ID:        1,
				Address:   "http://ya.ru",
				Shortened: "mocked",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s, _ := newTestStore(t)
			created, err := s.CreateResource(t.Context(), test.want)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.wantErr, err)
				return
			}
			require.NoError(t, err)
			got, err := s.GetResourceByURL(t.Context(), created.Shortened)
			require.NoError(t, err)
			assert.Equal(t, created, got)
		})
	}
}
