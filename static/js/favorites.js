document.querySelectorAll('.country-card').forEach(card => {
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