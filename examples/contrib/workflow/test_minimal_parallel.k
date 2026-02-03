// Minimal parallel executor test

let makeParallel = import "examples/contrib/workflow/parallel_executor.k"
let Parallel = makeParallel()

log("Creating parallel executor...")

let executor = Parallel.createParallelExecutor({
    workerCount: 2,
    queueSize: 10,
    batchSize: 5,
    enableMetrics: false,
})

log("Executor created!")
log("Config:", executor.config)
