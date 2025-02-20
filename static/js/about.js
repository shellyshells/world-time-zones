let slideIndex = 1;
showSlides(slideIndex);

function changeSlide(n) {
    showSlides(slideIndex += n);
}

function currentSlide(n) {
    showSlides(slideIndex = n);
}

function showSlides(n) {
    let slides = document.getElementsByClassName("slide");
    let dots = document.getElementsByClassName("dot");
    
    if (n > slides.length) {slideIndex = 1}
    if (n < 1) {slideIndex = slides.length}
    
    for (let i = 0; i < slides.length; i++) {
        slides[i].style.display = "none";
        dots[i].className = dots[i].className.replace(" active", "");
    }
    
    slides[slideIndex-1].style.display = "block";
    dots[slideIndex-1].className += " active";
}

// Updated fullscreen functionality
function openFullscreen(img) {
    const overlay = document.getElementById('fullscreenOverlay');
    const fullscreenImg = document.getElementById('fullscreenImage');
    
    fullscreenImg.src = img.src;
    fullscreenImg.alt = img.alt;
    overlay.style.display = 'block';
    
    document.body.style.overflow = 'hidden';
}

function closeFullscreen() {
    const overlay = document.getElementById('fullscreenOverlay');
    overlay.style.display = 'none';
    document.body.style.overflow = 'auto';
}

// Updated click handlers for fullscreen
document.getElementById('fullscreenOverlay').addEventListener('click', function() {
    closeFullscreen();
});

document.getElementById('fullscreenImage').addEventListener('click', function(e) {
    e.stopPropagation();  // Prevent overlay click when clicking image
    closeFullscreen();
});

// Close fullscreen with Escape key
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && document.getElementById('fullscreenOverlay').style.display === 'block') {
        closeFullscreen();
    }
});

// Add active class to current page in navigation
document.addEventListener('DOMContentLoaded', function() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-links a');
    
    navLinks.forEach(link => {
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
});