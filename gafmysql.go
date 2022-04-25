package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/rk0cc-xyz/gaf/structure"
	"github.com/rk0cc-xyz/gafmysql"
)

type PrintContext struct {
	LastUpdate string                                `json:"last_update"`
	Context    []structure.GitHubRepositoryStructure `json:"context"`
}

func main() {
	setmode := flag.Bool("set", false, "Archive current data of GitHub repository API")
	getmode := flag.Bool("get", false, "Get recent archived data of GitHub repository API")
	page := flag.Int64("page", 1, "Display which page of the result returned, default is `1` (`-get` only)")
	ppi := flag.Int64("ppi", 10, "How many repositories will be displayed in a single page, default is `10` (only accept 10 - 100 and can be divided by 10, `-get` only)")

	flag.Parse()

	if *setmode && !*getmode {
		serr := gafmysql.ArchiveCurrentAPIToDB()
		if serr != nil {
			panic(serr)
		}
	} else if *getmode && !*setmode {
		getContext(*page, *ppi)
	} else {
		flag.PrintDefaults()
	}
}

func rangedRepo(page int64, ppi int64) (*PrintContext, error) {
	ctx, lu, gerr := gafmysql.GetArchivedRepositoryAPI()
	if gerr != nil {
		return nil, gerr
	}

	start := int64(float64(page-1) * float64(ppi))
	end := int64(float64(page) * float64(ppi))

	if end > int64(len(ctx)) {
		end = int64(len(ctx))
	}

	return &PrintContext{
		Context:    ctx[start:end],
		LastUpdate: *lu,
	}, nil
}

func getContext(page int64, ppi int64) {
	if ppi%10 != 0 {
		panic("unaccepted page per items value: " + strconv.FormatInt(page, 10))
	}

	rp, rperr := rangedRepo(page, ppi)
	if rperr != nil {
		panic(rperr)
	}
	pj, pjerr := json.Marshal(rp)
	if pjerr != nil {
		printEmptyJson()
	}

	fmt.Println(string(pj))
}

func printEmptyJson() {
	fmt.Println("{}")
	os.Exit(0)
}
