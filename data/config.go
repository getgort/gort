/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package data

import "time"

// GortConfig is the top-level configuration object
type GortConfig struct {
	GortServerConfigs GortServerConfigs `yaml:"gort,omitempty"`
	GlobalConfigs     GlobalConfigs     `yaml:"global,omitempty"`
	DatabaseConfigs   DatabaseConfigs   `yaml:"database,omitempty"`
	DockerConfigs     DockerConfigs     `yaml:"docker,omitempty"`
	JaegerConfigs     JaegerConfigs     `yaml:"jaeger,omitempty"`
	KubernetesConfigs KubernetesConfigs `yaml:"kubernetes,omitempty"`
	SlackProviders    []SlackProvider   `yaml:"slack,omitempty"`
	DiscordProviders  []DiscordProvider `yaml:"discord,omitempty"`
}

// GortServerConfigs is the data wrapper for the "gort" section.
type GortServerConfigs struct {
	AllowSelfRegistration bool   `yaml:"allow_self_registration,omitempty"`
	APIAddress            string `yaml:"api_address,omitempty"`
	APIURLBase            string `yaml:"api_url_base,omitempty"`
	DevelopmentMode       bool   `yaml:"development_mode,omitempty"`
	EnableSpokenCommands  bool   `yaml:"enable_spoken_commands,omitempty"`
	TLSCertFile           string `yaml:"tls_cert_file,omitempty"`
	TLSKeyFile            string `yaml:"tls_key_file,omitempty"`
}

// GlobalConfigs is the data wrapper for the "global" section
type GlobalConfigs struct {
	CommandTimeout time.Duration `yaml:"command_timeout,omitempty"`
}

// DatabaseConfigs is the data wrapper for the "database" section.
type DatabaseConfigs struct {
	Host                  string        `yaml:"host,omitempty"`
	Port                  int           `yaml:"port,omitempty"`
	User                  string        `yaml:"user,omitempty"`
	Password              string        `yaml:"password,omitempty"`
	SSLEnabled            bool          `yaml:"ssl_enabled,omitempty"`
	ConnectionMaxIdleTime time.Duration `yaml:"connection_max_idle_time,omitempty"`
	ConnectionMaxLifetime time.Duration `yaml:"connection_max_life_time,omitempty"`
	MaxIdleConnections    int           `yaml:"max_idle_connections,omitempty"`
	MaxOpenConnections    int           `yaml:"max_open_connections,omitempty"`
	QueryTimeout          time.Duration `yaml:"query_timeout,omitempty"`
}

// DockerConfigs is the data wrapper for the "docker" section.
type DockerConfigs struct {
	DockerHost string `yaml:"host,omitempty"`
	Network    string `yaml:"network,omitempty"`
}

// JaegerConfigs is the data wrapper for the "jaeger" section.
type JaegerConfigs struct {
	Endpoint string `yaml:"endpoint,omitempty"`
	Password string `yaml:"password,omitempty"`
	Username string `yaml:"username,omitempty"`
}

// KubernetesConfigs is the data wrapper for the "kubernetes" section.
type KubernetesConfigs struct {
	Namespace             string `yaml:"namespace,omitempty"`
	EndpointFieldSelector string `yaml:"endpoint_field_selector,omitempty"`
	EndpointLabelSelector string `yaml:"endpoint_label_selector,omitempty"`
	PodFieldSelector      string `yaml:"pod_field_selector,omitempty"`
	PodLabelSelector      string `yaml:"pod_label_selector,omitempty"`
}
