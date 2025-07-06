package main

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/servers"
)

func main() {

	name, version := "yarumo-app", "1.0.0"
	ctx, options := context.Background(), GetOptions()
	boot.Run[core.Config](ctx, name, version, func(ctx context.Context, app servers.Application) error {
		wctx, err := boot.Context[core.Config]()
		if err != nil {
			return fmt.Errorf("error getting context: %w", err)
		}

		fmt.Println("Configuration:", fmt.Sprintf("%+v", wctx.Config))

		return nil
	}, options...)
}

/*
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mockRestClient := boot.Get[comm.RESTClient](&wctx.Container, "MockRestClient")
	resp, err := mockRestClient.Call(timeoutCtx, http.MethodGet, "", nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	if resp.Code != http.StatusOK {
		fmt.Println(fmt.Sprintf("Response err: %+v", resp)) //nolint:gosimple
	}

	rest := comm.NewRESTClient("https://fakerestapi.azurewebsites.net", comm.WithHTTPClient(wctx.HttpClient))
	resp, err = rest.Call(timeoutCtx, http.MethodGet, "/api/v1/Activities", nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	if pointer.IsSlice(resp.Data) {
		sliceMaps, err := comm.ToSliceOfMapsOfAny(resp.Data)
		if err != nil {
			return fmt.Errorf("error converting response data to map: %w", err)
		}
		fmt.Println(fmt.Sprintf("Response status: %+v", sliceMaps)) //nolint:gosimple
	}
	if pointer.IsMap(resp.Data) {
		maps, err := comm.ToMapOfAny(resp.Data)
		if err != nil {
			return fmt.Errorf("error converting response data to map: %w", err)
		}
		fmt.Println(fmt.Sprintf("Response status: %+v", maps)) //nolint:gosimple
	}
*/
