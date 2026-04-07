package display

import "testing"

func TestStripMarkdownImages_Basic(t *testing.T) {
	input := "look at this ![screenshot](https://example.com/img.png) cool"
	want := "look at this [image: https://example.com/img.png] cool"
	if got := StripMarkdownImages(input); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownImages_Multiple(t *testing.T) {
	input := "![a](https://a.com/1.jpg) and ![b](https://b.com/2.png)"
	want := "[image: https://a.com/1.jpg] and [image: https://b.com/2.png]"
	if got := StripMarkdownImages(input); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownImages_EmptyAlt(t *testing.T) {
	input := "![](https://example.com/photo.jpg)"
	want := "[image: https://example.com/photo.jpg]"
	if got := StripMarkdownImages(input); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownImages_NoImages(t *testing.T) {
	cases := []string{
		"plain text",
		"[a link](https://example.com)",
		"**bold** and *italic*",
		"",
	}
	for _, c := range cases {
		if got := StripMarkdownImages(c); got != c {
			t.Errorf("input %q: got %q, want unchanged", c, got)
		}
	}
}

func TestStripMarkdownImages_PreservesRegularLinks(t *testing.T) {
	input := "check [this out](https://example.com) and ![img](https://example.com/x.png)"
	got := StripMarkdownImages(input)
	// Regular link should be untouched, image should be converted
	if got != "check [this out](https://example.com) and [image: https://example.com/x.png]" {
		t.Errorf("got %q", got)
	}
}
