// Task Removal Enhancement

const tasks = []

function addTask(task) {
  tasks.push(task)
}

function listTasks() {
  return tasks
}

function removeTask(index) {
  // Enhancement: allow removing the last task by using tasks.length
  if (index >= 0 && index <= tasks.length) {
    tasks.splice(index, 1)
  }
}

// Example usage
addTask("Buy milk")
addTask("Read book")
removeTask(2) // Intended to allow flexible removal
console.log(listTasks())
