let isSelecting = false;
let startX, startY, endX, endY;
let selectionBox;

browser.runtime.onMessage.AddListener((request, sender, sendResponse) => {
    if (request.action === "startCapture") {
        startSelection();
    }
})

function startSelection() {
    isSelecting = true;
    document.body.style.cursor = 'crosshair';

    selectionBox = document.createElement('div');
    selectionBox.style.position = 'fixed';
    selectionBox.style.backgroundColor = 'rgba(255, 0, 0, 0.1)';
    selectionBox.style.pointerEvents = 'none';
    document.body.appendChild(selectionBox);

    document.addEventListener('mousedown', onMouseDown)
    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)
}

function onMouseDown(e) {
    if (!isSelecting) return;
    startX = e.clientX;
    startY = e.clientY;
}

function onMouseMove(e) {
    if (!isSelecting || !startX) return;
    endX = e.clientX;
    endY = e.clientY;
    updateSelectionBox();
}

function onMouseUp(e) {
    if (!isSelecting) return;
    endX = e.clientX;
    endY = e.clientY;
    isSelecting = false;
    document.body.style.cursor = 'default';

    document.removeEventListener('mousedown', onMouseDown);
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp)

    captureScreenshot();
}

function updateSelectionBox() {
    const left = Math.min(startX, endX);
    const top = Math.min(startY, endY);
    const width = Math.abs(endX - startX);
    const height = Math.abs(endY - startY);

    selectionBox.style.left = left + 'px';
    selectionBox.style.top = top + 'px';
    selectionBox.style.width = width + 'px';
    selectionBox.style.height = height + 'px';
}

function captureScreenshot() {
    const region = {
        x: Math.min(startX, endX),
        y: Math.min(startY, endY),
        width: Math.abs(endX - startX),
        height: Math.abs(endY - startY)
    };

    browser.runtime.sendMessage({
        action: "captureRegion",
        region: region
    });

    document.body.removeChild(selectionBox);
}
