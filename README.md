# caddy-dns-fetcher
Allows you to query a hostname over dns and use the returning value however you like in caddy.

Example:
```
{
  order dnsfetcher before redir
}
localhost {
  dnsfetcher TXT example.org
  respond "Value: {dnsfetcher.response}"
}
```