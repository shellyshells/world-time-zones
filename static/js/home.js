function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function submitForm() {
    const form = document.getElementById('searchForm');
    
    const urlParams = new URLSearchParams(window.location.search);
    const currentPage = urlParams.get('page');
    
    if (currentPage && currentPage !== '1') {
        const pageInput = document.createElement('input');
        pageInput.type = 'hidden';
        pageInput.name = 'page';
        pageInput.value = '1';
        form.appendChild(pageInput);
    }
    
    const formData = new FormData(form);
    const cleanForm = new FormData();
    
    for (let [key, value] of formData.entries()) {
        if (value.trim() !== '') {
            cleanForm.append(key, value);
        }
    }
    
    const params = new URLSearchParams(cleanForm);
    window.location.href = '?' + params.toString();
}

document.querySelectorAll('.country-card').forEach(card => {
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