# Instrument errors

This utility will parse all the go code outside of `/scripts` and instrument it
so that all returned errors contain a call trace (with the usage of
`tracerr.Wrap()`).  It also modifies all error checks to unwrap them for proper
comparison of the underlying error.

This instrumentation allows printing the call trace from the point at which
this error is first returned within this code base.
