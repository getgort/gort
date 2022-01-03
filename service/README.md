# service

This is where Gort's RESTful service lives.

# Decision Log

* 20220103 How to respond no "not found":
  * Single-object requests (i.e. `/v2/bundles/{name}/versions/{version}`) should respond with a 404 (Not Found)
  * List requests  (i.e. `/v2/bundles/{name}/versions`) should respond with a 204 (No Content)
* 20220103 "Exists" functionality provided by the service should be implemented using the HEAD method.