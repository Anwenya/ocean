package test

import (
	"github.com/blevesearch/bleve/v2"
	"testing"
)

func TestBleve(t *testing.T) {
	mapping := bleve.NewIndexMapping()
	bleve.New("", mapping)
}
