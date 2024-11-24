package main

import "net/http"

func mainHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	// for _, v := range req.Form {
	//     if v
	// }

	// res.Write([]byte("Привет!"))
	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("http://localhost:8080/EwHXdJfB"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
