from flask import Flask, request, jsonify

app = Flask(__name__)

# 存储所有任务的内存数据结构
todos = {}
current_id = 1

# 添加任务
@app.route('/tasks', methods=['POST'])
def add_task():
    global current_id
    data = request.get_json()
    if not data or 'title' not in data:
        return jsonify({"error": "Task title is required"}), 400
    
    task = {
        'id': current_id,
        'title': data['title'],
        'completed': False
    }
    todos[current_id] = task
    current_id += 1
    return jsonify(task), 201

# 查看所有任务
@app.route('/tasks', methods=['GET'])
def get_tasks():
    return jsonify(list(todos.values()))

# 更新任务状态
@app.route('/tasks/<int:task_id>', methods=['PUT'])
def update_task(task_id):
    data = request.get_json()
    if task_id not in todos:
        return jsonify({"error": "Task not found"}), 404

    task = todos[task_id]
    if 'title' in data:
        task['title'] = data['title']
    if 'completed' in data:
        task['completed'] = data['completed']
    
    return jsonify(task)

# 删除任务
@app.route('/tasks/<int:task_id>', methods=['DELETE'])
def delete_task(task_id):
    if task_id not in todos:
        return jsonify({"error": "Task not found"}), 404
    
    del todos[task_id]
    return jsonify({"message": "Task deleted successfully"})

if __name__ == '__main__':
    app.run(debug=True)
