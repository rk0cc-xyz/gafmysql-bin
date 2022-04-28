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

type PrintContextPaged struct {
	PrintContext
	HasPrev bool `json:"has_prev"`
	HasNext bool `json:"has_next"`
}

func main() {
	setmode := flag.Bool("set", false, "Archive current data of GitHub repository API")
	getmode := flag.Bool("get", false, "Get recent archived data of GitHub repository API")
	page := flag.Int64("page", 1, "Display which page of the result returned, default is `1` (`-get` only)")
	ppi := flag.Int64("ppi", 10, "How many repositories will be displayed in a single page, default is `10` (only accept 10 - 100 and can be divided by 10, `-get` only)")
	printall := flag.Bool("all", false, "Print all repositories data as once")

	flag.Parse()

	if *setmode && !*getmode {
		serr := gafmysql.ArchiveCurrentAPIToDB()
		if serr != nil {
			panic(serr)
		}
	} else if *getmode && !*setmode {
		if *printall {
			getAllContext()
		} else {
			getContext(*page, *ppi)
		}
	} else {
		flag.PrintDefaults()
	}
}

func load_context() ([]structure.GitHubRepositoryStructure, *string, error) {
	ctx, lu, gerr := gafmysql.GetArchivedRepositoryAPI()
	if gerr != nil {
		return nil, nil, gerr
	}

	return ctx, lu, nil
}

func rangedRepo(page int64, ppi int64) (*PrintContextPaged, error) {
	ctx, lu, gerr := load_context()
	if gerr != nil {
		return nil, gerr
	}

	if len(ctx) == 0 && page == 1 {
		return &PrintContextPaged{
			PrintContext: PrintContext{
				Context:    []structure.GitHubRepositoryStructure{},
				LastUpdate: *lu,
			},
			HasPrev: false,
			HasNext: false,
		}, nil
	}

	start := int64(float64(page-1) * float64(ppi))
	end := int64(float64(page) * float64(ppi))

	last_page := end > int64(len(ctx))

	if start > int64(len(ctx)) {
		return nil, nil
	} else if last_page {
		end = int64(len(ctx))
	}

	return &PrintContextPaged{
		PrintContext: PrintContext{
			Context:    ctx[start:end],
			LastUpdate: *lu,
		},
		HasPrev: page > 1,
		HasNext: !last_page,
	}, nil
}

func getAllContext() {
	ctx, lu, gerr := load_context()
	if gerr != nil {
		panic(gerr)
	}

	paj, pajerr := json.Marshal(PrintContext{
		Context:    ctx,
		LastUpdate: *lu,
	})

	if pajerr != nil {
		printEmptyJson()
	}

	fmt.Println(string(paj))
}

func getContext(page int64, ppi int64) {
	if ppi%10 != 0 {
		panic("unaccepted page per items value: " + strconv.FormatInt(page, 10))
	}

	rp, rperr := rangedRepo(page, ppi)
	if rperr != nil {
		panic(rperr)
	} else if rp == nil {
		printEmptyJson()
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
