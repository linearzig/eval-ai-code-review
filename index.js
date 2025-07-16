// Task Manager Example

const tasks = []

function addTask(task) {
  tasks.push(task)
}

function listTasks() {
  return tasks
}

function removeTask(index) {
  if (index >= 0 && index < tasks.length) {
    tasks.splice(index, 1)
  }
}

// Example usage
addTask("Buy milk")
addTask("Read book")
removeTask(0)
console.log(listTasks())
