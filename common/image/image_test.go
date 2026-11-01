package image_test

import (
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/jhopan/Panway/common/client"
	img "github.com/jhopan/Panway/common/image"

	"github.com/stretchr/testify/assert"
	_ "golang.org/x/image/webp"
)

type CountingReader struct {
	reader    io.Reader
	BytesRead int
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.BytesRead += n
	return n, err
}

var (
	cases = []struct {
		url    string
		format string
		width  int
		height int
	}{
		{"https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg", "jpeg", 2560, 1669},
		{"https://upload.wikimedia.org/wikipedia/commons/9/97/Basshunter_live_performances.png", "png", 4500, 2592},
		{"https://upload.wikimedia.org/wikipedia/commons/c/c6/TO_THE_ONE_SOMETHINGNESS.webp", "webp", 984, 985},
		{"https://upload.wikimedia.org/wikipedia/commons/d/d0/01_Das_Sandberg-Modell.gif", "gif", 1917, 1533},
		{"https://upload.wikimedia.org/wikipedia/commons/6/62/102Cervus.jpg", "jpeg", 270, 230},
	}
)

// fetchFixture mengunduh fixture gambar.
//
// Wikimedia menolak request tanpa User-Agent browser (HTTP 400), jadi test
// di-skip kalau fixture tidak tersedia — bukan menggagalkan seluruh suite
// karena gangguan jaringan pihak ketiga.
func fetchFixture(t *testing.T, url string) io.Reader {
	t.Helper()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Skipf("fixture request gagal: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Panway/1.0; +https://github.com/jhopan/Panway)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("fixture tidak bisa diambil: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Skipf("fixture tidak tersedia: HTTP %d", resp.StatusCode)
	}
	return resp.Body
}

func TestMain(m *testing.M) {
	client.Init()
	m.Run()
}

func TestDecode(t *testing.T) {
	// Bytes read: varies sometimes
	// jpeg: 1063892
	// png: 294462
	// webp: 99529
	// gif: 956153
	// jpeg#01: 32805
	for _, c := range cases {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			reader := &CountingReader{reader: fetchFixture(t, c.url)}
			decoded, format, err := image.Decode(reader)
			assert.NoError(t, err)
			size := decoded.Bounds().Size()
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, size.X)
			assert.Equal(t, c.height, size.Y)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}

	// Bytes read:
	// jpeg: 4096
	// png: 4096
	// webp: 4096
	// gif: 4096
	// jpeg#01: 4096
	for _, c := range cases {
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			reader := &CountingReader{reader: fetchFixture(t, c.url)}
			config, format, err := image.DecodeConfig(reader)
			assert.NoError(t, err)
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, config.Width)
			assert.Equal(t, c.height, config.Height)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}
}

func TestBase64(t *testing.T) {
	// Bytes read:
	// jpeg: 1063892
	// png: 294462
	// webp: 99072
	// gif: 953856
	// jpeg#01: 32805
	for _, c := range cases {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			data, err := io.ReadAll(fetchFixture(t, c.url))
			assert.NoError(t, err)
			encoded := base64.StdEncoding.EncodeToString(data)
			reader := &CountingReader{reader: base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))}
			decoded, format, err := image.Decode(reader)
			assert.NoError(t, err)
			size := decoded.Bounds().Size()
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, size.X)
			assert.Equal(t, c.height, size.Y)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}

	// Bytes read:
	// jpeg: 1536
	// png: 768
	// webp: 768
	// gif: 1536
	// jpeg#01: 3840
	for _, c := range cases {
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			data, err := io.ReadAll(fetchFixture(t, c.url))
			assert.NoError(t, err)
			encoded := base64.StdEncoding.EncodeToString(data)
			reader := &CountingReader{reader: base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))}
			config, format, err := image.DecodeConfig(reader)
			assert.NoError(t, err)
			assert.Equal(t, c.format, format)
			assert.Equal(t, c.width, config.Width)
			assert.Equal(t, c.height, config.Height)
			t.Logf("Bytes read: %d", reader.BytesRead)
		})
	}
}

func TestGetImageSize(t *testing.T) {
	for i, c := range cases {
		t.Run("Decode:"+strconv.Itoa(i), func(t *testing.T) {
			// Pre-check the fixture: some thumbnail URLs are no longer
			// served by Wikimedia (HTTP 400), and GetImageSize would
			// then report zero instead of skipping.
			fetchFixture(t, c.url)
			width, height, err := img.GetImageSize(c.url)
			assert.NoError(t, err)
			assert.Equal(t, c.width, width)
			assert.Equal(t, c.height, height)
		})
	}
}

func TestGetImageSizeFromBase64(t *testing.T) {
	for i, c := range cases {
		t.Run("Decode:"+strconv.Itoa(i), func(t *testing.T) {
			data, err := io.ReadAll(fetchFixture(t, c.url))
			assert.NoError(t, err)
			encoded := base64.StdEncoding.EncodeToString(data)
			width, height, err := img.GetImageSizeFromBase64(encoded)
			assert.NoError(t, err)
			assert.Equal(t, c.width, width)
			assert.Equal(t, c.height, height)
		})
	}
}
