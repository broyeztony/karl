// Minimal test to reproduce parallel executor deadlock

let makeParallel = import "examples/contrib/workflow/parallel_executor.k"
let Parallel = makeParallel()

log("Creating parallel executor...")

let executor = Parallel.createParallelExecutor({
    workerCount: 2,
    queueSize: 10,
    batchSize: 5,
    enablePriority: false,
    shutdownTimeout: 30000,
    enableMetrics: false,
})

log("Executor created!")

// Create 3 simple tasks
let tasks = [
    {
        name: "Task-1",
        handler: (ctx) -> {
            log("[TASK-1] Running")
            sleep(50)
            { success: true, data: "Result 1" }
        }
    },
    {
        name: "Task-2",
        handler: (ctx) -> {
            log("[TASK-2] Running")
            sleep(50)
            { success: true, data: "Result 2" }
        }
    },
    {
        name: "Task-3",
        handler: (ctx) -> {
            log("[TASK-3] Running")
            sleep(50)
            { success: true, data: "Result 3" }
        }
    },
]

log("Executing", tasks.length, "tasks...")
let result = executor.execute(tasks, {})

log("Result:", result)
log("Success:", result.success)
