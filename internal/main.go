package main

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/guidomantilla/yarumo/pkg/pointer"
	"github.com/guidomantilla/yarumo/pkg/utils"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	_ = lo.Empty[int]()
	s := utils.RandomString(100)
	fmt.Println(s)

	s = utils.Substring(s, 10, 20)
	fmt.Println(s)

	fmt.Println(utils.ChunkString(s, 3))
	fmt.Println(pointer.IsType(&s, "string"))
	fmt.Println(utils.Words("Hello, world! This is a test string."))
	fmt.Println(utils.Chunk([]string{"123", "123", "123", "123", "123", "123", "123", "123", "123", "123", "123", "123"}, 5))
	fmt.Println(utils.Delete(0, []string{"123", "123"}))
	fmt.Println(utils.DeleteRange(0, 1, []string{"123", "123"}))
}
