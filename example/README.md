To run the example, `cd` into this directory and run:

```bash
./run.sh
```

This will start a local cluster of three nodes and an API endpoint. Three consecutive requests will be made to the API endpoint, and the results will be printed to the console.

In this example, Singleflight will be invoked to ensure that only one data load is performed for the requested key on the mock slow DB.
