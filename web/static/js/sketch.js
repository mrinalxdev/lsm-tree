let memTable = [];
let sstables = [];
let operations = [];

function setup() {
    const canvas = createCanvas(600, 400);
    canvas.parent('canvasContainer');
    frameRate(30);
}

function draw() {
    background(255);
    drawMemTable();
    drawSSTables();
    drawOperations();
}

function drawMemTable() {
    push();
    translate(50, 50);

    fill(200, 220, 255);
    stroke(100, 140, 255);
    rect(0, 0, 200, 150);

    fill(0);
    noStroke();
    textAlign(CENTER);
    text("MemTable", 100, -10);

    textAlign(LEFT);
    let y = 20;
    for (let entry of memTable) {
        text(`${entry.key}: ${entry.value}`, 10, y);
        y += 20;
    }
    pop();
}

function drawSSTables() {
    push();
    translate(300, 50);
    for (let i = 0; i < sstables.length; i++) {
        fill(220, 255, 220);
        stroke(140, 255, 140);
        rect(0, i * 80, 250, 60);

        fill(0);
        noStroke();
        textAlign(CENTER);
        text(`SSTable ${i + 1}`, 125, i * 80 - 10);
        textAlign(LEFT);
        let y = i * 80 + 20;
        for (let entry of sstables[i]) {
            text(`${entry.key}: ${entry.value}`, 10, y);
            y += 20;
        }
    }
    pop();
}

function drawOperations() {
    push();
    translate(50, 250);
    fill(255, 220, 220);
    stroke(255, 140, 140);
    rect(0, 0, 500, 100);

    fill(0);
    noStroke();
    textAlign(CENTER);
    text("Recent Operations", 250, -10);
    textAlign(LEFT);
    let y = 20;
    for (let op of operations.slice(-4)) {
        text(op, 10, y);
        y += 20;
    }
    pop();
}

function addOperation(type, key, value = null) {
    const timestamp = new Date().toLocaleTimeString();
    const op = value
        ? `${timestamp} - ${type}: ${key} = ${value}`
        : `${timestamp} - ${type}: ${key}`;
    operations.push(op);

    if (operations.length > 10) {
        operations.shift();
    }
}

function handleWSMessage(data) {
    switch (data.type) {
        case 'set':
            memTable.push({ key: data.key, value: data.value });
            addOperation('SET', data.key, data.value);
            break;

        case 'get':
            addOperation('GET', data.key, data.value);
            break;

        case 'delete':
            memTable = memTable.filter(entry => entry.key !== data.key);
            addOperation('DELETE', data.key);
            break;

        case 'flush':
            sstables.unshift([...memTable]);
            memTable = [];
            addOperation('FLUSH', 'MemTable');
            break;
    }
}

function updateStatus(message, isError = false) {
    const statusElement = document.getElementById('operationStatus');
    if (statusElement) {
        statusElement.textContent = message;
        statusElement.className = `mt-2 text-sm ${isError ? 'text-red-500' : 'text-green-500'}`;
        setTimeout(() => {
            statusElement.textContent = '';
        }, 3000);
    }
}
