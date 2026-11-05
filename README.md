# EVM RPC Agent

EVM RPC Agent sits between applications and an EVM-compatible RPC node. It has two main features:

- Crawl blockchain data from the RPC node and store it in a database.
- Serve RPC requests from the database first, and pass a request through to the RPC node when the data is not found or the request is not supported.

## License

EVM RPC Agent is licensed under GPL-3.0 license. See COPYING file for details.
