package assetLoader

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestQueryIndex(t *testing.T) {
	tests := []struct {
		query   string
		want    int
		wantErr bool
	}{
		{query: "", want: 0},
		{query: "?frame=7", want: 7},
		{query: "?frame=-1", wantErr: true},
		{query: "?frame=one", wantErr: true},
	}
	gin.SetMode(gin.TestMode)
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest("GET", "/preview/example.dc6"+test.query, nil)
		got, err := queryIndex(context, "frame")
		if (err != nil) != test.wantErr {
			t.Fatalf("query %q error = %v, wantErr %v", test.query, err, test.wantErr)
		}
		if !test.wantErr && got != test.want {
			t.Fatalf("query %q value = %d, want %d", test.query, got, test.want)
		}
	}
}
