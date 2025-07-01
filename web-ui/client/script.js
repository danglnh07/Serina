// Piece mapping for image paths
const pieceMapping = {
    'P': 'w-pawn', 'R': 'w-rook', 'N': 'w-knight', 'B': 'w-bishop', 'Q': 'w-queen', 'K': 'w-king',
    'p': 'b-pawn', 'r': 'b-rook', 'n': 'b-knight', 'b': 'b-bishop', 'q': 'b-queen', 'k': 'b-king',
    ' ': null
};

// Create chessboard
const chessboard = document.getElementById('chessboard');
const files = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'];
const ranks = [8, 7, 6, 5, 4, 3, 2, 1];
const squareIds = []; //All 64 squares' ID (h5, g6, ...)
let selectedPiece = null;
let currentPosition = []; //Current board state
let validMoves = []; // All valid moves for ONE piece (selected piece)
let currentMoves = []; // Store current moves (all moves) from the latest API response
let sideToMove = 'white'; // Track side to move
let moveIndex = 0; // Track move history index

// Generate board squares
ranks.forEach(rank => {
    files.forEach(file => {
        // Create square
        const squareId = `${file}${rank}`;
        squareIds.push(squareId);
        const square = document.createElement('div');
        square.id = squareId;
        square.classList.add('square');

        // Set light or dark square
        const isLight = (files.indexOf(file) + rank) % 2 === 0;
        square.classList.add(isLight ? 'light' : 'dark');

        // Add file label (for rank 1)
        if (rank === 1) {
            const fileLabel = document.createElement('span');
            fileLabel.textContent = file;
            fileLabel.classList.add('label', 'file-label');
            square.appendChild(fileLabel);
        }

        // Add rank label (for file h)
        if (file === 'h') {
            const rankLabel = document.createElement('span');
            rankLabel.textContent = rank;
            rankLabel.classList.add('label', 'rank-label');
            square.appendChild(rankLabel);
        }

        // Add click handler for moving to highlighted squares
        square.addEventListener('click', handleSquareClick);

        // Add square to the chessboard
        chessboard.appendChild(square);
    });
});

// Function to update board with position from array
function updateBoard(chessData) {
    // Set the current position to the parameter array
    currentPosition = chessData.board;

    // Loop through each square and add/remove piece according to the current position array
    squareIds.forEach((squareId, index) => {
        const square = document.getElementById(squareId);

        // Remove existing piece
        const existingPiece = square.querySelector('.piece');
        if (existingPiece) {
            existingPiece.remove();
        }

        // Add new piece if present
        const piece = currentPosition[index];
        if (piece && pieceMapping[piece]) {
            // Create piece's image
            const img = document.createElement('img');
            img.src = `./assets/${pieceMapping[piece]}.png`;
            img.classList.add('piece');
            img.dataset.piece = piece; // Store piece type
            img.dataset.square = squareId; // Store square ID

            // Handle piece selection
            img.addEventListener('click', handlePieceClick);

            // Handle clicks on pieces for captures (forward to square's click handler)
            img.addEventListener('click', (event) => {
                event.stopPropagation(); // Prevent square click from firing
                handleSquareClick({ currentTarget: square }); // Trigger square's click handler
            });

            // Add image to the square
            square.appendChild(img);
        }
    });

    // Update side to move
    sideToMove = chessData.side_to_move;

    // Update the list of current moves
    currentMoves = chessData.moves;
}

// Function to clear highlights
function clearHighlights() {
    squareIds.forEach(squareId => {
        const square = document.getElementById(squareId);
        square.classList.remove('highlight');
    });
}

// Function to highlight squares a piece can move to 
function highlightMoves(moves, piece, fromSquare) {
    clearHighlights();

    // Filter all squares that 'piece' can move to
    validMoves = moves.filter(move => {
        // For castling, only allow if the piece is a king
        if (['o-o', 'o-o-o', 'O-O', 'O-O-O'].includes(move) && (piece === 'K' || piece === 'k')) {
            // Check if castling move matches piece color (White: O-O/O-O-O, Black: o-o/o-o-o)
            return (piece === 'K' && move.startsWith('O')) || (piece === 'k' && move.startsWith('o'));
        }

        // For regular moves (including promotion), check if the move starts from the selected square
        if (move.length >= 4) {
            return move.slice(0, 2) === fromSquare;
        }


        return false;
    });

    // For each valid moves, we found the target square (destination square) and add highlight to them
    validMoves.forEach(move => {
        let targetSquare;

        // Calculate target square
        if (move === 'O-O' || move === 'o-o') { // King-side castling: king moves two squares right
            targetSquare = piece === 'K' ? 'g1' : 'g8';
        } else if (move === 'O-O-O' || move === 'o-o-o') { // Queen-side castling: king moves two squares left
            targetSquare = piece === 'K' ? 'c1' : 'c8';
        } else { // Regular move: target is the last two characters
            targetSquare = move.slice(2, 4);
        }

        // Get the target square from DOM and highlight it
        const square = document.getElementById(targetSquare);
        if (square) {
            square.classList.add('highlight');
        }
    });
}

// Handle piece click (for piece selection)
function handlePieceClick(event) {
    // Get the piece and the square it's located
    const piece = event.target.dataset.piece;
    const squareId = event.target.dataset.square;

    // Check if piece belongs to the current side to move
    const isWhitePiece = /[PRNBQK]/.test(piece); // Uppercase for White
    const isBlackPiece = /[prnbqk]/.test(piece); // Lowercase for Black
    if ((sideToMove === 'white' && !isWhitePiece) || (sideToMove === 'black' && !isBlackPiece)) {
        return; // Ignore clicks on opponent pieces
    }


    if (selectedPiece === event.target) { // Clicking the same piece again: release it
        clearHighlights();
        selectedPiece = null;
        validMoves = [];
    } else { // Select new piece
        selectedPiece = event.target;
        highlightMoves(currentMoves, piece, squareId);
    }
}

// Handle square click for moving
async function handleSquareClick(event) {
    // If no piece selected, stop here
    if (!selectedPiece) return;

    // Get the from, destination square of the move and the piece moving
    const targetSquare = event.currentTarget;
    const fromSquare = selectedPiece.dataset.square;
    const piece = selectedPiece.dataset.piece;

    if (targetSquare.id === fromSquare) {
        return;
    }

    // Find the move in validMoves
    let move = null;

    // Check for castling moves
    if (piece === 'K' || piece === 'k') {
        const kingSideCastle = piece === 'K' ? 'O-O' : 'o-o';
        const queenSideCastle = piece === 'K' ? 'O-O-O' : 'o-o-o';
        if (validMoves.includes(kingSideCastle) && targetSquare.id === (piece === 'K' ? 'g1' : 'g8')) {
            move = kingSideCastle;
        } else if (validMoves.includes(queenSideCastle) && targetSquare.id === (piece === 'K' ? 'c1' : 'c8')) {
            move = queenSideCastle;
        }
    }

    // Check for regular moves
    if (!move) {
        const potentialMove = `${fromSquare}${targetSquare.id}`;
        const movesFound = validMoves.filter(mv => {
            return mv.startsWith(potentialMove);
        });

        if (movesFound.length === 0) {
            selectedPiece = null;
            clearHighlights();
            return;
        } else if (movesFound.length === 1) {
            move = potentialMove;
        } else {
            // If this is a promotion (white pawn at rank 7 or black pawn at rank 2), prompt for promotion option
            if (piece === 'P' && fromSquare.slice(1) === '7') {
                let promotionOption = prompt('Enter promotiton option (Q, R, N, B ):');
                if (['Q', 'R', 'N', 'B'].includes(promotionOption)) {
                    move = potentialMove + promotionOption;
                } else {
                    return;
                }
            }

            if (piece === 'p' && fromSquare.slice(1) === '2') {
                let promotionOption = prompt('Enter promotiton option (q, r, n, b ):');
                if (['q', 'r', 'n', 'b'].includes(promotionOption)) {
                    move = potentialMove + promotionOption;
                } else {
                    return;
                }
            }
        }
    }

    // If move is found (valid move), perform move
    if (move) {
        // Make request to API endpoint: /move?move=YOUR_MOVE
        let url = `http://localhost:8080/move?move=${encodeURIComponent(move)}`
        fetchPosition(url);

        // Insert move into history table
        const moveNumber = Math.floor(moveIndex / 2) + 1;
        const moveHtml = `
            <span class="history-move" data-id="${moveIndex}">
              <img src="./assets/${pieceMapping[piece]}.png" class="history-piece">
              ${move}
            </span>
          `;
        if (sideToMove === 'white') {
            // White's move: create new row
            const row = document.createElement('tr');
            row.innerHTML = `
              <th scope="row">${moveNumber}</th>
              <td>${moveHtml}</td>
              <td></td>
            `;
            historyBody.appendChild(row);
        } else if (sideToMove === 'black') {
            // Black's move: update last row
            const lastRow = historyBody.lastElementChild;
            if (lastRow) {
                lastRow.cells[2].innerHTML = moveHtml;
            }
        }
        moveIndex++;
    }
}

// Function to fetch position from API 
// Any method that make change to the current state of the board can use this function
async function fetchPosition(url) {
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Failed to fetch position');
        }
        const data = await response.json();
        updateBoard(data);
        clearHighlights(); // Clear any existing highlights
        selectedPiece = null; // Reset selected piece
        // Clear history table
        // historyBody.innerHTML = '';
        // moveIndex = 0;
    } catch (error) {
        console.error('Error fetching position:', error);
        alert('Failed to load position. Please try again.');
    }
}


// Load default position on page load
window.addEventListener('load', () => {
    fetchPosition('http://localhost:8080/fen');
});

// Handle FEN button click
document.getElementById('fen-btn').addEventListener('click', () => {
    const fen = prompt('Enter FEN string:');
    if (fen !== null) {
        fetchPosition(`http://localhost:8080/fen?fen=${encodeURIComponent(fen)}`);
    }
});

// Handle Flip button click
document.getElementById('flip-btn').addEventListener('click', () => {
    //ranks.reverse();
    fetchPosition('http://localhost:8080/flip');
});

// Function to fetch Perft results
async function fetchPerft(depth) {
    try {
        // Show modal with loading spinner
        const perftModal = new bootstrap.Modal(document.getElementById('perftModal'));
        document.getElementById('perftLoading').style.display = 'flex';
        document.getElementById('modal-content').style.display = 'none';
        // document.getElementById('totalNodes').textContent = '';
        // document.getElementById('time').textContent = '';
        perftModal.show();

        const response = await fetch(`http://localhost:8080/perft?depth=${depth}`);
        if (!response.ok) {
            throw new Error('Failed to fetch Perft results');
        }
        const data = await response.json();
        displayPerftResults(data);
    } catch (error) {
        console.error('Error fetching Perft results:', error);
        alert('Failed to load Perft results. Please try again.');
        const perftModal = bootstrap.Modal.getInstance(document.getElementById('perftModal'));
        perftModal.hide();
    }
}

// Initialize modal once
const modalElement = document.getElementById('resultsModal');
let resultsModal = null;
if (modalElement) {
    resultsModal = new bootstrap.Modal(modalElement);
} else {
    console.error('Results modal element not found');
}

// Generalized function to fetch data (Perft, Search, etc.)
async function fetchData(type, params = {}) {
    try {
        if (!resultsModal) throw new Error('Results modal not initialized');

        // Show spinner, clear content, and update modal title
        const modalSpinner = document.getElementById('modal-spinner');
        const modalContent = document.getElementById('modal-content');
        const modalTitle = document.getElementById('modal-title');
        modalSpinner.style.display = 'flex';
        modalContent.innerHTML = '';

        // Map type to title and endpoint
        const typeConfig = {
            perft: {
                title: 'Perft Results',
                url: `http://localhost:8080/perft?depth=${params.depth || 0}`,
                display: displayPerftResults
            },
            search: {
                title: 'Search Results',
                url: `http://localhost:8080/search?depth=${params.depth || 0}`,
                display: displaySearchResults
            }
        };

        if (!typeConfig[type]) {
            throw new Error(`Unknown data type: ${type}`);
        }

        modalTitle.textContent = typeConfig[type].title;
        resultsModal.show();

        const response = await fetch(typeConfig[type].url);
        if (!response.ok) {
            throw new Error(`Failed to fetch ${type} results`);
        }
        const data = await response.json();
        typeConfig[type].display(data);
    } catch (error) {
        console.error(`Error fetching ${type} results:`, error);
        alert(`Failed to load ${type} results. Please try again.`);
        if (resultsModal) {
            resultsModal.hide();
        }
    }
}

// Display Perft results
function displayPerftResults(data) {
    const modalContent = document.getElementById('modal-content');
    let html = `
        <table class="table table-striped">
            <thead>
                <tr>
                    <th>Move</th>
                    <th>Nodes</th>
                </tr>
            </thead>
            <tbody>
    `;
    for (const [move, nodes] of Object.entries(data.result)) {
        html += `<tr><td>${move}</td><td>${nodes.toLocaleString()}</td></tr>`;
    }
    html += `
            </tbody>
        </table>
        <p><strong>Total Nodes:</strong> ${data.total_node.toLocaleString()}</p>
        <p><strong>Time:</strong> ${data.time.toLocaleString()} ms</p>
    `;
    modalContent.innerHTML = html;
    document.getElementById('modal-spinner').style.display = 'none';
    if (resultsModal) {
        resultsModal.show();
    }
}

// Display Search results
function displaySearchResults(data) {
    const modalContent = document.getElementById('modal-content');
    const html = `
          <p><strong>Optimal Move:</strong> ${data.searched_move}</p>
          <p><strong>Time:</strong> ${data.time.toLocaleString()} ms</p>
        `;
    modalContent.innerHTML = html;
    document.getElementById('modal-spinner').style.display = 'none';
    if (resultsModal) {
        resultsModal.show();
    }
}


// Handle Perft button click
document.getElementById('perft-btn').addEventListener('click', () => {
    const depth = prompt('Enter depth (integer >= 0):');
    if (depth !== null) {
        const depthNum = parseInt(depth);
        if (isNaN(depthNum) || depthNum < 0) {
            alert('Please enter a valid non-negative integer.');
            return;
        }
        fetchData('perft', { depth: depthNum });
    }
});

document.getElementById('search-btn').addEventListener('click', () => {
    let depthVal = document.getElementById("depth").value;
    fetchData('search', { depth: depthVal });

});