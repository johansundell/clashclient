package main

import (
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	//router.Handle("/tmpl/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "./"}))
	//router.PathPrefix("/bower_components/").Handler(http.StripPrefix("/bower_components/", http.FileServer(http.Dir("bower_components"))))
	//router.PathPrefix("/tmpl/").Handler(http.StripPrefix("/tmpl/", http.FileServer(http.Dir("tmpl"))))
	router.PathPrefix("/bower_components/").Handler(http.StripPrefix("/bower_components/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "bower_components"})))
	router.PathPrefix("/tmpl/").Handler(http.StripPrefix("/tmpl/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "tmpl"})))
	//router.
	return router
}
