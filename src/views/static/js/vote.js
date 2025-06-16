document.addEventListener('DOMContentLoaded', function() {
    // Fonction pour afficher les messages
    function showMessage(message, isError = false) {
        const messageElement = document.getElementById('message');
        messageElement.textContent = message;
        messageElement.className = isError ? 'error' : 'success';
        messageElement.style.display = 'block';
        
        // Ajouter l'animation de sortie avant de cacher le message
        setTimeout(() => {
            messageElement.style.animation = 'slideOut 0.3s ease-out forwards';
            setTimeout(() => {
                messageElement.style.display = 'none';
                messageElement.style.animation = '';
            }, 300);
        }, 2700);
    }

    // Fonction pour mettre à jour l'état des boutons
    function updateVoteButtons(postId, value) {
        const upvoteBtn = document.querySelector(`.vote-btn.upvote[data-post-id="${postId}"]`);
        const downvoteBtn = document.querySelector(`.vote-btn.downvote[data-post-id="${postId}"]`);
        
        // Réinitialiser l'état des boutons
        upvoteBtn.classList.remove('active');
        downvoteBtn.classList.remove('active');
        
        // Mettre à jour l'état du bouton actif
        if (value === 1) {
            upvoteBtn.classList.add('active');
        } else if (value === -1) {
            downvoteBtn.classList.add('active');
        }
    }

    const voteButtons = document.querySelectorAll('.vote-btn');
    
    voteButtons.forEach(button => {
        button.addEventListener('click', async function() {
            const postId = parseInt(this.dataset.postId);
            const value = parseInt(this.dataset.value);
            const voteElement = document.getElementById(`votes-${postId}`);
            const originalScore = voteElement.textContent;
            
            try {
                const response = await fetch('/vote', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        post_id: postId,
                        value: value
                    })
                });

                const data = await response.json();
                
                if (data.error) {
                    showMessage(data.error, true);
                    if (data.redirect) {
                        window.location.href = data.redirect;
                    }
                    // Restaurer le score original en cas d'erreur
                    voteElement.textContent = originalScore;
                    return;
                }

                if (data.success) {
                    if (voteElement) {
                        voteElement.textContent = data.score;
                        updateVoteButtons(postId, value);
                    }
                }
            } catch (error) {
                console.error('Erreur lors du vote:', error);
                showMessage('Une erreur est survenue lors du vote', true);
                // Restaurer le score original en cas d'erreur
                voteElement.textContent = originalScore;
            }
        });
    });
}); 