// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// appengine provides a Google App Engine (Standard) specific implementation of myip
package main // import "bramp.net/myip/appengine"

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"bramp.net/myip/lib/conf"
	"bramp.net/myip/lib/myip"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/appengine"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var debugConfig = &conf.Config{
	Debug: true,

	Host:  "localhost:8080",
	Host4: "ip4-localhost.bramp.net:8080", // "127.0.0.1:8080",
	Host6: "ip6-localhost.bramp.net:8080", // "[::1]:8080",

	MapsAPIKey: "AIzaSyA6-HIkxuJEX6Hf3rzVx07no32YM3N5V9s",

	DisallowedHeaders: []string{"none"},
}

var prodConfig = &conf.Config{
	Debug: false,

	Host:  "ip.bramp.net",
	Host4: "ip4.bramp.net",
	Host6: "ip6.bramp.net",

	MapsAPIKey: "AIzaSyA6-HIkxuJEX6Hf3rzVx07no32YM3N5V9s",

	// If behind CloudFlare use the following:
	//IPHeader: "Cf-Connecting-Ip",
	//RequestIDHeader: "Cf-Ray",
}

var appengineDefaultConfig = &conf.Config{
	RequestIDHeader: "X-Cloud-Trace-Context",
	LatLongHeader:   "X-Appengine-Citylatlong",
	CityHeader:      "X-Appengine-City",

	DisallowedHeaders: []string{
		"X-Appengine-Default-Namespace",
		"X-Appengine-Request-Id-Hash",
		"X-Appengine-Request-Log-Id",
		"X-Appengine-Default-Version-Hostname",
		"X-Appengine-User-Email",
		"X-Appengine-User-Id",
		"X-Appengine-User-Is-Admin",
		"X-Appengine-User-Nickname",
		"X-Appengine-User-Organization",
		"X-Appengine-Server-Name",
		"X-Appengine-Server-Port",
		"X-Appengine-Server-Protocol",
		"X-Appengine-Server-Software",
		"X-Appengine-Remote-Addr",

		"X-Cloud-Trace-Context",
		"X-Google-Apps-Metadata",
		"X-Zoo",

		"Cf-Connecting-Ip",
		"Cf-Ipcountry",
		"Cf-Ray",
		"Cf-Visitor",
	},
}

func main() {
	r := mux.NewRouter()

	// IsAppEngine tests if running on AppEngine
	if appengine.IsAppEngine() {
		// ProxyHeaders takes X-Forwarded Headers and populates the request with this information.
		r.Use(handlers.ProxyHeaders)

		// Warmup handler
		r.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
			log.Println("warmup done")
		})
	}

	config := config()

	myip.Register(r, config)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	s := &http.Server{
		Addr: ":" + port,

		// Log all requests using the standard Apache format.
		// TODO Ensure this is following the AppEngine best practices
		Handler: handlers.CombinedLoggingHandler(os.Stderr, r),
	}

	log.Printf("Listening on port %s for %s", port, config.Host)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe() failed: %s:", err)
	}
}

func config() *conf.Config {
	config := debugConfig
	log.SetLevel(log.DebugLevel)

	if appengine.IsAppEngine() {
		config = prodConfig
		log.SetLevel(log.WarnLevel)

		var err error
		config, err = conf.ApplyDefaults(config, appengineDefaultConfig)
		if err != nil {
			log.Fatalf("Failed to ApplyDefaults: %s", err)
		}

		// Load the MapsAPISigningKey secret key
		loadSecrets(config)
	}

	config.Version = Version
	config.BuildTime = BuildTime

	return config
}

// Some keys are stored in GCP Secret Manager.
func loadSecrets(config *conf.Config) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Errorf("failed to setup secretmanager client: %v", err)
		return
	}
	defer client.Close()

	// GCP project in which to store secrets in Secret Manager.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	keyBytes, err := accessSecret(ctx, client, projectID, "map_static_signing")
	if err != nil {
		log.Errorf("failed to access secret: %v", err)
		return
	}

	key, err := base64.URLEncoding.DecodeString(string(keyBytes))
	if err != nil {
		log.Errorf("failed to decode the MapsAPISigningKey: %v", err)
		return
	}

	config.MapsAPISigningKey = key
}

func accessSecret(ctx context.Context, client *secretmanager.Client, projectID, secretID string) ([]byte, error) {
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID),
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, err
	}

	return result.Payload.Data, nil
}
