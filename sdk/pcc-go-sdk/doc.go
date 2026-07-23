/*
Package pccclient loads strongly typed configuration from PCC.

Create a reusable client with a default namespace and configuration:

	client, err := pccclient.New(
		endpoint,
		accessKey,
		secret,
		pccclient.WithNamespace("production"),
		pccclient.WithConfig("application"),
	)

Then decode the complete document or a single field:

	var config ApplicationConfig
	err = client.Load(&config)

	var port int
	err = client.LoadField("server.port", &port)

Clients are safe for concurrent reads as their configuration is immutable after
construction.
*/
package pccclient
