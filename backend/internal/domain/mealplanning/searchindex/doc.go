/*
Package searchindex holds what meal planning puts into a search index: the subset of each
indexed type a document carries, the conversions that build one, and the index names.

They are domain types rather than part of the indexing service because the manager reads
them back out of an index, and the domain does not import services.
*/
package searchindex
