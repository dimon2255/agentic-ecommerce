package admin

import "testing"

func TestUploadURLRequest_Valid(t *testing.T) {
	req := UploadURLRequest{
		Filename:    "product-photo.png",
		ContentType: "image/png",
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestUploadURLRequest_AllowedContentTypes(t *testing.T) {
	types := []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	for _, ct := range types {
		req := UploadURLRequest{Filename: "test.img", ContentType: ct}
		if err := req.Validate(); err != nil {
			t.Errorf("expected %s to be allowed, got error: %v", ct, err)
		}
	}
}

func TestUploadURLRequest_InvalidContentType(t *testing.T) {
	invalid := []string{"application/javascript", "text/html", "application/octet-stream", "image/svg+xml"}
	for _, ct := range invalid {
		req := UploadURLRequest{Filename: "test.img", ContentType: ct}
		if err := req.Validate(); err == nil {
			t.Errorf("expected %s to be rejected", ct)
		}
	}
}

func TestUploadURLRequest_PathTraversalFilename(t *testing.T) {
	filenames := []string{
		"../../../etc/passwd",
		"..\\windows\\system32",
		"foo/bar.png",
		"foo\\bar.png",
		"test\x00.png",
	}
	for _, fn := range filenames {
		req := UploadURLRequest{Filename: fn, ContentType: "image/png"}
		if err := req.Validate(); err == nil {
			t.Errorf("expected filename %q to be rejected", fn)
		}
	}
}

func TestUploadURLRequest_EmptyFields(t *testing.T) {
	req := UploadURLRequest{}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty fields")
	}
}

func TestUploadURLRequest_FilenameTooLong(t *testing.T) {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	req := UploadURLRequest{Filename: string(longName), ContentType: "image/png"}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for long filename")
	}
}
