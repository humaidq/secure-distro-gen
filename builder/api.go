package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"git.sr.ht/~humaid/linux-gen/config"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("key")
	if key != config.Config.SecretKey {
		fmt.Println("fail!")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Forbidden")
		return
	}
	fmt.Println("Success!")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		config.Logger.Error(err)
		return
	}

	var cust Customisation
	err = json.Unmarshal(body, &cust)
	if err != nil {
		config.Logger.Error(err)
		return
	}
	fmt.Println("Got request for " + cust.AuthorID + ". Building...")

	fmt.Fprintf(w, "OK")
	w.WriteHeader(200)

	go func() {
		_, err = Start(cust)
		if err != nil {
			config.Logger.Error(err)
			return
		}
	}()
}

func RunAPI() {
	http.HandleFunc("/api", apiHandler)
	err := http.ListenAndServe("0.0.0.0:8484", nil)
	if err != nil {
		panic(err)
	}
}
