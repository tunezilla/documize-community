// Copyright 2016 Documize Inc. <legal@documize.com>. All rights reserved.
//
// This software (Documize Community Edition) is licensed under
// GNU AGPL v3 http://www.gnu.org/licenses/agpl-3.0.en.html
//
// You can operate outside the AGPL restrictions by purchasing
// Documize Enterprise Edition and obtaining a commercial license
// by contacting <sales@documize.com>.
//
// https://documize.com

// Package env provides runtime, server level setup and configuration
package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// LoadConfig loads runtime parameters like port numbers and DB connections.
// We first check for -config switch that would point us towards a .CONF file.
// If not found, we then read parameters from command line and environment vars.
func LoadConfig() (f Flags, ok bool) {
	// Check and process config file
	f, ok = configFile()

	// If not OK then get parameters from command line and environment variables.
	if !ok {
		f, ok = commandLineEnv()
	}

	// reserved
	if len(f.Location) == 0 {
		f.Location = "selfhost"
	}

	if len(f.Bucket) == 0 {
		f.Bucket = "community"
	}

	return
}

// configFile checks for the presence of zero or one command argument.
// If no arguments are provided then we look for and load documize.conf file.
// If one argument is provided then we load the specified config file.
// If more than one argument is provided then we exit as flags as have passed.
// checks to see if it is a TOML format config file.
func configFile() (f Flags, ok bool) {
	ok = false
	var configFile string

	// First argument is always program being executed.
	// No additional arguments means check for documize.conf file.
	if len(os.Args) == 1 {
		// No arguments, so we default to default config filename.
		configFile = "documize.conf"
	} else if len(os.Args) == 2 {
		// Config filename passed in, so we use it.
		configFile = os.Args[1]
	} else {
		// Too many arguments means flags passed in so we return.
		return
	}

	// Does file exist?
	if len(configFile) == 0 || !configFileExists(configFile) {
		return
	}

	// Tell caller where the config came from.
	f.ConfigSource = configFile

	// We parse the TOML format config file.
	var ct ConfigToml
	if _, err := toml.DecodeFile(configFile, &ct); err != nil {
		fmt.Println(err)
		return
	}
	f.DBType = strings.ToLower(ct.Database.Type)
	f.DBConn = ct.Database.Connection
	f.Salt = ct.Database.Salt
	f.HTTPPort = strconv.Itoa(ct.HTTP.Port)
	f.ForceHTTPPort2SSL = strconv.Itoa(ct.HTTP.ForceSSLPort)
	f.SSLCertFile = ct.HTTP.Cert
	f.SSLKeyFile = ct.HTTP.Key
	f.TLSVersion = ct.HTTP.TLSVersion
	f.Location = strings.ToLower(ct.Install.Location)

	if len(f.TLSVersion) == 0 {
		f.TLSVersion = "1.3"
	}

	ok = true
	return
}

// commandLineEnv loads command line and OS environment variables required by the program to function.
func commandLineEnv() (f Flags, ok bool) {
	ok = true
	var dbConn, dbType, dbMaxIdleConns, dbMaxOpenConns, dbConnMaxLifetime, dbConnMaxIdleTime, jwtKey, siteMode, port, certFile, keyFile, forcePort2SSL, TLSVersion, bucket, minioEndpoint, location string

	// register(&configFile, "salt", false, "the salt string used to encode JWT tokens, if not set a random value will be generated")
	register(&jwtKey, "salt", false, "the salt string used to encode JWT tokens, if not set a random value will be generated")
	register(&certFile, "cert", false, "the cert.pem file used for https")
	register(&keyFile, "key", false, "the key.pem file used for https")
	register(&port, "port", false, "http/https port number")
	register(&forcePort2SSL, "forcesslport", false, "redirect given http port number to TLS")
	register(&TLSVersion, "tlsversion", false, "select minimum TLS: 1.0, 1.1, 1.2, 1.3")
	register(&siteMode, "offline", false, "set to '1' for OFFLINE mode")
	register(&dbType, "dbtype", true, "specify the database provider: mysql|percona|mariadb|postgresql|sqlserver")
	register(&dbConn, "db", true, `'database specific connection string for example "user:password@tcp(localhost:3306)/dbname"`)
	register(&dbMaxIdleConns, "dbmaxidleconns", false, `maximum idle connections`)
	register(&dbMaxOpenConns, "dbmaxopenconns", false, `maximum open connections`)
	register(&dbConnMaxLifetime, "dbconnmaxlifetime", false, `maximum connection lifetime in seconds`)
	register(&dbConnMaxIdleTime, "dbconnmaxidletime", false, `maximum connection idle time in seconds`)
	register(&bucket, "bucket", false, `object storage bucket, defaults to "community"`)
	register(&minioEndpoint, "minio", false, `specify to use minio instead of S3, for example "http://id:secret@minio:9000"`)
	register(&location, "location", false, `reserved`)

	if !parse("db") {
		ok = false
	}

	f.DBType = strings.ToLower(dbType)
	f.DBConn = dbConn
	f.DBMaxIdleConns = dbMaxIdleConns
	f.DBMaxOpenConns = dbMaxOpenConns
	f.ForceHTTPPort2SSL = forcePort2SSL
	f.HTTPPort = port
	f.Salt = jwtKey
	f.SiteMode = siteMode
	f.SSLCertFile = certFile
	f.SSLKeyFile = keyFile
	f.TLSVersion = TLSVersion
	f.Bucket = bucket
	f.MinioEndpoint = minioEndpoint
	f.Location = strings.ToLower(location)
	f.ConfigSource = "flags/environment"

	if len(f.TLSVersion) == 0 {
		f.TLSVersion = "1.2"
	}

	if len(f.DBMaxIdleConns) == 0 {
		f.DBMaxIdleConns = "30"
	}

	if len(f.DBMaxOpenConns) == 0 {
		f.DBMaxOpenConns = "100"
	}

	if len(f.DBConnMaxLifetime) == 0 {
		f.DBConnMaxLifetime = "14400"
	}

	if len(f.DBConnMaxIdleTime) == 0 {
		f.DBConnMaxIdleTime = "0"
	}

	return f, ok
}

func configFileExists(fn string) bool {
	info, err := os.Stat(fn)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
