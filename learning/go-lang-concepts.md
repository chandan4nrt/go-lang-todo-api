In Go, a Context is a way to manage the lifecycle of a request. It tells a function (like a database query) how long it is allowed to run and when it should stop.

Cursor (mongo.Cursor) - A Cursor is a pointer to the result set of a query. Instead of loading 10,000 documents into your RAM all at once, MongoDB sends them in batches.

The "Why": It’s a memory-saving mechanism. A cursor allows you to "stream" results from the database.

Iterative Processing: You can use cursor.Next(ctx) to pull one document at a time, or cursor.All(ctx, &results) to pull everything into a slice (if you know the list is small).

Resource Cleanup: Cursors hold an open connection to the database. You must call cursor.Close(ctx) when you are done, or you will eventually run out of database connections (a "connection leak").