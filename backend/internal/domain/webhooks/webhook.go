/*
Package webhooks is this application's half of platform-go's webhooks: the data
change events a webhook write emits.

The model is platform's — an endpoint, and the subscriptions under it — and so
are the store, the dispatcher and the gRPC surface; see
internal/repositories/postgres/webhooksstore and internal/build/webhooks. What
is not platform's is which events exist, which is generated into catalog from
the constants every domain declares.
*/
package webhooks
