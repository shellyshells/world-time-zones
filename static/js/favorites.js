const favoriteStates = new Map();

        function toggleFavorite(event, countryName) {
            event.stopPropagation();
            const btn = event.target;
            updateFavoriteUI(countryName, 'remove');
            updateFavoriteAPI(countryName, 'remove');
        }

        function handleDoubleClick(event, countryName) {
            const card = event.currentTarget;
            const favoriteBtn = card.querySelector('.favorite-btn');
            updateFavoriteUI(countryName, 'remove');
            updateFavoriteAPI(countryName, 'remove');
        }

        function updateFavoriteUI(countryName, action) {
            const buttons = document.querySelectorAll(`.country-card[data-country="${countryName}"] .favorite-btn`);
            buttons.forEach(btn => {
                btn.textContent = action === 'add' ? '★' : '☆';
                btn.dataset.favorited = action === 'add' ? 'true' : 'false';
                
                btn.classList.add('updating');
                setTimeout(() => btn.classList.remove('updating'), 300);
            });

            favoriteStates.set(countryName, action === 'add');
        }

        function updateFavoriteAPI(countryName, action) {
            fetch('/api/favorite', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    country: countryName,
                    action: action
                })
            })
            .then(response => {
                if (response.ok) {
                    const cards = document.querySelectorAll(`.country-card[data-country="${countryName}"]`);
                    cards.forEach(card => {
                        card.style.opacity = '0';
                        card.style.transition = 'opacity 0.3s ease-out';
                        setTimeout(() => {
                            card.remove();
                            if (document.querySelectorAll('.country-card').length === 0) {
                                location.reload();
                            }
                        }, 300);
                    });
                }
            })
            .catch(error => {
                console.error('Error:', error);
                const revertAction = action === 'add' ? 'remove' : 'add';
                updateFavoriteUI(countryName, revertAction);
            });
        }

        document.querySelectorAll('.country-card').forEach(card => {
            // Add double click handler
            card.setAttribute('ondblclick', `handleDoubleClick(event, '${card.dataset.country}')`);
            
            // Single click for flip
            card.addEventListener('click', (e) => {
                // Don't flip if clicking the favorite button
                if (!e.target.classList.contains('favorite-btn') && !e.target.closest('.favorite-btn')) {
                    card.classList.toggle('is-flipped');
                }
            });
        });

        function switchTab(event, tabId) {
            event.stopPropagation(); // Prevent card from flipping when clicking tabs
        
            // Get the container of the clicked tab
            const tabContainer = event.target.closest('.tab-container');
        
            // Remove active class from all tabs and contents in this container
            tabContainer.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
            tabContainer.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
        
            // Add active class to clicked tab and corresponding content
            event.target.classList.add('active');
            document.getElementById(tabId).classList.add('active');
        }