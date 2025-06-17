document.addEventListener('DOMContentLoaded', function() {
    const navbarSearch = document.querySelector('.navbar-search');
    
    navbarSearch.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            const query = this.value.trim();
            if (query) {
                window.location.href = `/search?query=${encodeURIComponent(query)}&type=text`;
            }
        }
    });
}); 