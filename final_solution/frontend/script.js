// Base authentication credentials
const AUTH_CREDENTIALS = "Basic " + btoa("bal:2fourall");

// API base URL
const API_BASE_URL = "";

// Object to store virtual services data
let virtualServices = [];
let certificates = [];
let blacklistedIPs = [];

// Auto refresh interval (in milliseconds)
const AUTO_REFRESH_INTERVAL = 30000; // 30 seconds
let autoRefreshTimer;
let isLoading = false; // Changed to false as default - no loading animation

// Document ready event
document.addEventListener("DOMContentLoaded", () => {
    // Initialize tabs with smooth transitions
    initializeTabs();
    
    // Set up event handlers
    setupEventHandlers();
    
    // Load initial data
    loadDashboardData().then(() => {
        // Start auto refresh
        startAutoRefresh();
        
        // Set last updated time
        updateLastUpdatedTime();
    });
});

// Initialize tabs with smooth transitions
function initializeTabs() {
    const triggerTabList = [].slice.call(document.querySelectorAll('a[data-bs-toggle="tab"]'));
    triggerTabList.forEach(triggerEl => {
        triggerEl.addEventListener('click', function (event) {
            event.preventDefault();
            const targetTab = document.querySelector(this.getAttribute('href'));
            const activeTab = document.querySelector('.tab-pane.active');
            
            // Remove active class from all tabs
            document.querySelectorAll('.nav-link').forEach(el => el.classList.remove('active'));
            document.querySelectorAll('.tab-pane').forEach(el => {
                el.classList.remove('show', 'active');
            });
            
            // Add active class to clicked tab
            this.classList.add('active');
            
            // Animate the tab transition
            activeTab.style.opacity = 0;
            setTimeout(() => {
                activeTab.classList.remove('show', 'active');
                targetTab.classList.add('show', 'active');
                
                // Fade in the new tab
                targetTab.style.opacity = 0;
                setTimeout(() => {
                    targetTab.style.opacity = 1;
                }, 50);
                
                // Load data for the selected tab if needed
                loadTabContent(targetTab.id);
            }, 200);
        });
    });
}

// Load content based on selected tab
function loadTabContent(tabId) {
    switch (tabId) {
        case 'virtual-services-content':
            loadVirtualServices();
            break;
        case 'certificates-content':
            loadCertificates();
            break;
        case 'ip-rules-content':
            displayBlacklistedIPs();
            break;
        case 'rate-limiting-content':
            displayRateLimits();
            break;
        case 'content-routing-content':
            setupContentRoutingTab();
            break;
    }
}

// Start auto refresh timer
function startAutoRefresh() {
    if (autoRefreshTimer) {
        clearInterval(autoRefreshTimer);
    }
    
    autoRefreshTimer = setInterval(() => {
        // Only refresh if not currently loading data
        if (!isLoading) {
            const activeTabId = document.querySelector('.tab-pane.active').id;
            
            // Show a mini loading spinner in the refresh button
            const refreshBtn = document.getElementById('refresh-dashboard');
            if (refreshBtn) {
                refreshBtn.innerHTML = '<i class="fas fa-sync-alt fa-spin me-1"></i> Refreshing...';
                refreshBtn.disabled = true;
            }
            
            // Refresh the active tab content
            loadTabContent(activeTabId);
            
            // Update last updated time
            updateLastUpdatedTime();
            
            // Restore refresh button after a delay
            setTimeout(() => {
                if (refreshBtn) {
                    refreshBtn.innerHTML = '<i class="fas fa-sync-alt me-1"></i> Refresh';
                    refreshBtn.disabled = false;
                }
            }, 1000);
        }
    }, AUTO_REFRESH_INTERVAL);
}

// Update the "Last Updated" timestamp
function updateLastUpdatedTime() {
    const lastUpdated = document.getElementById('last-updated');
    if (lastUpdated) {
        const now = new Date();
        lastUpdated.textContent = `Updated: ${now.toLocaleTimeString()}`;
        
        // Animate the update
        lastUpdated.classList.add('bg-primary');
        setTimeout(() => {
            lastUpdated.classList.remove('bg-primary');
            lastUpdated.classList.add('bg-secondary');
        }, 1000);
    }
}

// Show loading overlay
function showLoadingOverlay() {
    isLoading = true;
    // Function is now empty - no loading overlay 
}

// Hide loading overlay
function hideLoadingOverlay() {
    isLoading = false;
    // Function is now empty - no loading overlay
}

// Setup event handlers
function setupEventHandlers() {
    // Add server button in VS form
    document.getElementById('add-server-btn').addEventListener('click', addServerToForm);
    
    // Save VS button
    document.getElementById('save-vs-btn').addEventListener('click', saveVirtualService);
    
    // Generate certificate button
    document.getElementById('generate-cert-btn').addEventListener('click', generateCertificate);
    
    // Block IP button
    document.getElementById('block-ip-btn').addEventListener('click', blockIP);
    
    // Update rate limit button
    document.getElementById('update-rate-limit-btn').addEventListener('click', updateRateLimit);
    
    // Virtual service selector change (for content routing)
    document.getElementById('vs-selector').addEventListener('change', loadContentRoutingRules);
    
    // Save content routing rule button
    document.getElementById('save-rule-btn').addEventListener('click', saveContentRoutingRule);
    
    // Refresh dashboard button
    document.getElementById('refresh-dashboard').addEventListener('click', function() {
        // Animate the button
        this.innerHTML = '<i class="fas fa-sync-alt fa-spin me-1"></i> Refreshing...';
        this.disabled = true;
        
        // Reload dashboard data
        loadDashboardData().then(() => {
            // Restore button state after a delay
            setTimeout(() => {
                this.innerHTML = '<i class="fas fa-sync-alt me-1"></i> Refresh';
                this.disabled = false;
                
                // Update last updated time
                updateLastUpdatedTime();
            }, 1000);
        });
    });
    
    // VS Algorithm dropdown change
    document.getElementById('vs-algorithm').addEventListener('change', function() {
        // Show/hide relevant fields based on algorithm
        const isContentBased = this.value === 'content_based';
        const rateFields = document.querySelectorAll('#vs-form .rate-limit-field');
        
        if (isContentBased) {
            showAlert('When using content-based routing, you must add routing rules after creating the virtual service.', 'info');
        }
    });
    
    // Setup form animations
    setupFormAnimations();
}

// Setup form animations
function setupFormAnimations() {
    // Focus animations for form inputs
    document.querySelectorAll('.form-control, .form-select').forEach(input => {
        input.addEventListener('focus', function() {
            const label = this.previousElementSibling;
            if (label && label.classList.contains('form-label')) {
                label.classList.add('text-primary');
            }
        });
        
        input.addEventListener('blur', function() {
            const label = this.previousElementSibling;
            if (label && label.classList.contains('form-label')) {
                label.classList.remove('text-primary');
            }
        });
    });
    
    // Modal animations
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('show.bs.modal', function() {
            // Reset form fields and remove validation classes
            const form = this.querySelector('form');
            if (form) {
                form.reset();
                form.querySelectorAll('.is-invalid').forEach(el => {
                    el.classList.remove('is-invalid');
                });
            }
        });
    });
}

// Load dashboard data
async function loadDashboardData() {
    try {
        // Use Promise.all to load data in parallel
        await Promise.all([
            loadVirtualServices(),
            loadCertificates()
        ]);
        
        updateDashboardStats();
        return true;
    } catch (error) {
        console.error('Error loading dashboard data:', error);
        showAlert('Failed to load dashboard data. Check console for details.', 'danger');
        return false;
    } finally {
        // Set loading to false 
        isLoading = false;
    }
}

// Function to create a fetch with timeout
function fetchWithTimeout(url, options, timeout = 10000) {
    return Promise.race([
        fetch(url, options),
        new Promise((_, reject) => 
            setTimeout(() => reject(new Error('Request timed out')), timeout)
        )
    ]);
}

// Load virtual services
async function loadVirtualServices() {
    try {
        const response = await fetchWithTimeout(`${API_BASE_URL}/access/vs`, {
            headers: {
                'Authorization': AUTH_CREDENTIALS
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        virtualServices = await response.json();
        
        // Animate the count change in stats
        animateNumberChange('vs-count', virtualServices.length);
        
        displayVirtualServices();
        updateHealthTable();
        return virtualServices;
    } catch (error) {
        console.error('Error loading virtual services:', error);
        showAlert('Failed to load virtual services: ' + error.message, 'danger');
        virtualServices = [];
        displayVirtualServices();
        updateHealthTable();
        return [];
    }
}

// Load certificates
async function loadCertificates() {
    try {
        const response = await fetchWithTimeout(`${API_BASE_URL}/access/vs/certificates`, {
            headers: {
                'Authorization': AUTH_CREDENTIALS
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        certificates = await response.json();
        
        // Animate the count change in stats
        animateNumberChange('cert-count', Object.keys(certificates).length);
        
        displayCertificates();
        return certificates;
    } catch (error) {
        console.error('Error loading certificates:', error);
        showAlert('Failed to load certificates: ' + error.message, 'danger');
        certificates = {};
        displayCertificates();
        return {};
    }
}

// Animate number change in stats
function animateNumberChange(elementId, newValue) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const currentValue = parseInt(element.textContent) || 0;
    if (currentValue === newValue) return;
    
    // Highlight the element
    element.style.transition = 'color 0.3s ease';
    element.style.color = newValue > currentValue ? '#28a745' : '#dc3545';
    
    // Animate the number change
    let steps = 10;
    let step = 0;
    let increment = (newValue - currentValue) / steps;
    
    const animate = () => {
        step++;
        element.textContent = Math.round(currentValue + (increment * step));
        
        if (step < steps) {
            requestAnimationFrame(animate);
        } else {
            element.textContent = newValue;
            // Reset color after animation
            setTimeout(() => {
                element.style.color = '';
            }, 500);
        }
    };
    
    animate();
}

// Update dashboard stats
function updateDashboardStats() {
    // Count healthy servers
    let healthyServerCount = 0;
    virtualServices.forEach(vs => {
        if (vs.serverList) {
            vs.serverList.forEach(server => {
                if (server.health) {
                    healthyServerCount++;
                }
            });
        }
    });
    
    // Animate healthy servers count
    animateNumberChange('healthy-servers', healthyServerCount);
}

// Display virtual services
function displayVirtualServices() {
    const tbody = document.getElementById('vs-table').querySelector('tbody');
    
    // Save current scroll position
    const scrollPos = tbody.scrollTop;
    
    // Clear table with fade-out animation
    tbody.style.opacity = '0.5';
    
    setTimeout(() => {
        tbody.innerHTML = '';
        
        // Check if we have any virtual services
        if (virtualServices.length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = '<td colspan="6" class="text-center">No virtual services configured.</td>';
            tbody.appendChild(tr);
            tbody.style.opacity = '1';
            return;
        }
        
        virtualServices.forEach(vs => {
            const tr = document.createElement('tr');
            
            // Format algorithm name for display
            const algorithmDisplay = vs.algorithm.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
            
            // Check if VS has SSL certificate
            const hasSSL = certificates[vs.port] ? true : false;
            
            // Create row content
            tr.innerHTML = `
                <td>${vs.port}</td>
                <td>${algorithmDisplay}</td>
                <td>${vs.serverList ? vs.serverList.length : 0} servers</td>
                <td>${vs.rate_limit || 'None'}</td>
                <td>${hasSSL ? '<span class="badge bg-success">Yes</span>' : '<span class="badge bg-secondary">No</span>'}</td>
                <td class="actions">
                    <button class="btn btn-sm btn-info view-vs-btn" data-port="${vs.port}">
                        <i class="fas fa-eye"></i>
                    </button>
                    <button class="btn btn-sm btn-warning edit-vs-btn" data-port="${vs.port}">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-sm btn-danger delete-vs-btn" data-port="${vs.port}">
                        <i class="fas fa-trash"></i>
                    </button>
                </td>
            `;
            
            // Add with fade-in animation
            tr.style.opacity = '0';
            tbody.appendChild(tr);
            setTimeout(() => {
                tr.style.opacity = '1';
            }, 50);
        });
        
        // Restore table opacity
        setTimeout(() => {
            tbody.style.opacity = '1';
            
            // Restore scroll position
            tbody.scrollTop = scrollPos;
            
            // Add event listeners for action buttons
            addVSTableEventListeners();
            
            // Update related elements
            updateVSSelectors();
        }, 100);
    }, 200);
}

// Add event listeners to VS table buttons
function addVSTableEventListeners() {
    document.querySelectorAll('.view-vs-btn').forEach(btn => {
        btn.addEventListener('click', () => viewVirtualService(btn.dataset.port));
    });
    
    document.querySelectorAll('.edit-vs-btn').forEach(btn => {
        btn.addEventListener('click', () => editVirtualService(btn.dataset.port));
    });
    
    document.querySelectorAll('.delete-vs-btn').forEach(btn => {
        btn.addEventListener('click', () => deleteVirtualService(btn.dataset.port));
    });
}

// Update VS selectors in other forms
function updateVSSelectors() {
    // Update vs dropdown in content routing tab
    const vsSelector = document.getElementById('vs-selector');
    vsSelector.innerHTML = '<option value="">Select a virtual service</option>';
    
    virtualServices.forEach(vs => {
        const option = document.createElement('option');
        option.value = vs.port;
        option.textContent = `Port ${vs.port} (${vs.algorithm.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())})`;
        vsSelector.appendChild(option);
    });
    
    // Update port dropdown in certificate form
    const certPortSelector = document.getElementById('cert-port');
    certPortSelector.innerHTML = '';
    
    virtualServices.forEach(vs => {
        if (!certificates[vs.port]) {
            const option = document.createElement('option');
            option.value = vs.port;
            option.textContent = `Port ${vs.port}`;
            certPortSelector.appendChild(option);
        }
    });
    
    // Update select in rate limit table
    displayRateLimits();
}

// Display certificates
function displayCertificates() {
    const tbody = document.getElementById('cert-table').querySelector('tbody');
    
    // Save current scroll position
    const scrollPos = tbody.scrollTop;
    
    // Clear table with fade-out animation
    tbody.style.opacity = '0.5';
    
    setTimeout(() => {
        tbody.innerHTML = '';
        
        for (const [port, cert] of Object.entries(certificates)) {
            const tr = document.createElement('tr');
            
            tr.innerHTML = `
                <td>${port}</td>
                <td class="text-truncate">${cert.certPath}</td>
                <td class="text-truncate">${cert.keyPath}</td>
                <td>
                    <button class="btn btn-sm btn-warning renew-cert-btn" data-port="${port}">
                        <i class="fas fa-sync-alt me-1"></i> Renew
                    </button>
                </td>
            `;
            
            // Add with fade-in animation
            tr.style.opacity = '0';
            tbody.appendChild(tr);
            setTimeout(() => {
                tr.style.opacity = '1';
            }, 50);
        }
        
        // Add message if no certificates
        if (Object.keys(certificates).length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = '<td colspan="4" class="text-center">No SSL certificates configured.</td>';
            tbody.appendChild(tr);
        }
        
        // Restore table opacity
        setTimeout(() => {
            tbody.style.opacity = '1';
            
            // Restore scroll position
            tbody.scrollTop = scrollPos;
            
            // Add event listeners for renew button
            document.querySelectorAll('.renew-cert-btn').forEach(btn => {
                btn.addEventListener('click', () => renewCertificate(btn.dataset.port));
            });
        }, 100);
    }, 200);
}

// Update health table
function updateHealthTable() {
    const tbody = document.getElementById('health-table').querySelector('tbody');
    
    // Save current scroll position
    const scrollPos = tbody.scrollTop;
    
    // Clear table with fade-out animation
    tbody.style.opacity = '0.5';
    
    setTimeout(() => {
        tbody.innerHTML = '';
        
        virtualServices.forEach(vs => {
            const tr = document.createElement('tr');
            
            // Count healthy servers
            let healthyCount = 0;
            if (vs.serverList) {
                vs.serverList.forEach(server => {
                    if (server.health) {
                        healthyCount++;
                    }
                });
            }
            
            // Determine overall status
            let statusBadge;
            if (vs.serverList && vs.serverList.length > 0) {
                if (healthyCount === 0) {
                    statusBadge = '<span class="badge bg-danger">All Down</span>';
                } else if (healthyCount < vs.serverList.length) {
                    statusBadge = '<span class="badge bg-warning text-dark">Partially Healthy</span>';
                } else {
                    statusBadge = '<span class="badge bg-success">All Healthy</span>';
                }
            } else {
                statusBadge = '<span class="badge bg-secondary">No Servers</span>';
            }
            
            // Format algorithm name for display
            const algorithmDisplay = vs.algorithm.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
            
            tr.innerHTML = `
                <td>Port ${vs.port}</td>
                <td>${algorithmDisplay}</td>
                <td>${healthyCount}/${vs.serverList ? vs.serverList.length : 0} healthy</td>
                <td>${statusBadge}</td>
            `;
            
            // Add with fade-in animation
            tr.style.opacity = '0';
            tbody.appendChild(tr);
            setTimeout(() => {
                tr.style.opacity = '1';
            }, 50);
        });
        
        // Add message if no virtual services
        if (virtualServices.length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = '<td colspan="4" class="text-center">No virtual services configured.</td>';
            tbody.appendChild(tr);
        }
        
        // Restore table opacity
        setTimeout(() => {
            tbody.style.opacity = '1';
            
            // Restore scroll position
            tbody.scrollTop = scrollPos;
        }, 100);
    }, 200);
}

// Display blacklisted IPs
function displayBlacklistedIPs() {
    const tbody = document.getElementById('ip-table').querySelector('tbody');
    
    // Save current scroll position
    const scrollPos = tbody.scrollTop;
    
    // Clear table with fade-out animation
    tbody.style.opacity = '0.5';
    
    setTimeout(() => {
        tbody.innerHTML = '';
        
        blacklistedIPs.forEach(ip => {
            const tr = document.createElement('tr');
            
            tr.innerHTML = `
                <td>${ip.ip}</td>
                <td>${ip.timestamp}</td>
                <td>
                    <button class="btn btn-sm btn-secondary unblock-ip-btn" data-ip="${ip.ip}" disabled>
                        <i class="fas fa-unlock me-1"></i> Unblock
                    </button>
                </td>
            `;
            
            // Add with fade-in animation
            tr.style.opacity = '0';
            tbody.appendChild(tr);
            setTimeout(() => {
                tr.style.opacity = '1';
            }, 50);
        });
        
        // Add note if no IPs are blocked
        if (blacklistedIPs.length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = '<td colspan="3" class="text-center">No IP addresses are currently blocked.</td>';
            tbody.appendChild(tr);
        }
        
        // Restore table opacity
        setTimeout(() => {
            tbody.style.opacity = '1';
            
            // Restore scroll position
            tbody.scrollTop = scrollPos;
        }, 100);
    }, 200);
}

// Display rate limits
function displayRateLimits() {
    const tbody = document.getElementById('rate-limit-table').querySelector('tbody');
    
    // Save current scroll position
    const scrollPos = tbody.scrollTop;
    
    // Clear table with fade-out animation
    tbody.style.opacity = '0.5';
    
    setTimeout(() => {
        tbody.innerHTML = '';
        
        virtualServices.forEach(vs => {
            const tr = document.createElement('tr');
            
            tr.innerHTML = `
                <td>${vs.port}</td>
                <td>${vs.rate_limit || 'Not set'}</td>
                <td>${vs.status_code || '-'}</td>
                <td>${vs.message || '-'}</td>
                <td>
                    <button class="btn btn-sm btn-warning edit-rate-limit-btn" data-port="${vs.port}">
                        <i class="fas fa-edit me-1"></i> Edit
                    </button>
                </td>
            `;
            
            // Add with fade-in animation
            tr.style.opacity = '0';
            tbody.appendChild(tr);
            setTimeout(() => {
                tr.style.opacity = '1';
            }, 50);
        });
        
        // Add message if no virtual services
        if (virtualServices.length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = '<td colspan="5" class="text-center">No virtual services configured.</td>';
            tbody.appendChild(tr);
        }
        
        // Restore table opacity
        setTimeout(() => {
            tbody.style.opacity = '1';
            
            // Restore scroll position
            tbody.scrollTop = scrollPos;
            
            // Add event listeners for edit button
            document.querySelectorAll('.edit-rate-limit-btn').forEach(btn => {
                btn.addEventListener('click', () => editRateLimit(btn.dataset.port));
            });
        }, 100);
    }, 200);
}

// Setup content routing tab
function setupContentRoutingTab() {
    // Clear rules table
    document.getElementById('rules-table').querySelector('tbody').innerHTML = '';
    
    // Disable add rule button initially
    document.getElementById('add-rule-btn').disabled = true;
    
    // Check if a VS is selected
    const selectedVS = document.getElementById('vs-selector').value;
    if (selectedVS) {
        loadContentRoutingRules(selectedVS);
    } else {
        const tbody = document.getElementById('rules-table').querySelector('tbody');
        const tr = document.createElement('tr');
        tr.innerHTML = '<td colspan="5" class="text-center">Please select a virtual service to manage routing rules.</td>';
        tbody.appendChild(tr);
    }
}

// Load content routing rules for the selected virtual service
async function loadContentRoutingRules() {
    const vsSelector = document.getElementById('vs-selector');
    const selectedVS = vsSelector.value;
    
    if (!selectedVS) {
        return;
    }
    
    try {
        // Show a loading indicator in the rules table
        const tbody = document.getElementById('rules-table').querySelector('tbody');
        tbody.innerHTML = '<tr><td colspan="5" class="text-center"><div class="loading-spinner my-3"></div></td></tr>';
        
        const response = await fetchWithTimeout(`${API_BASE_URL}/access/vs/${selectedVS}/rules`, {
            headers: {
                'Authorization': AUTH_CREDENTIALS
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const rules = await response.json();
        displayContentRoutingRules(rules, selectedVS);
        
        // Enable add rule button
        document.getElementById('add-rule-btn').disabled = false;
        
        // Set the current VS port in the add rule form
        document.getElementById('rule-vs-port').value = selectedVS;
        
        // Populate server dropdown in the add rule form
        const vs = virtualServices.find(v => v.Port == selectedVS);
        const serverSelector = document.getElementById('rule-server-name');
        serverSelector.innerHTML = '';
        
        if (vs && vs.ServerList) {
            vs.ServerList.forEach(server => {
                const option = document.createElement('option');
                option.value = server.Name;
                option.textContent = `${server.Name} (${server.URL})`;
                serverSelector.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading content routing rules:', error);
        showAlert('Failed to load content routing rules. Check console for details.', 'danger');
        
        // Show error in the rules table
        const tbody = document.getElementById('rules-table').querySelector('tbody');
        tbody.innerHTML = '<tr><td colspan="5" class="text-center text-danger">Error loading rules. Please try again.</td></tr>';
    }
}

// Display content routing rules
function displayContentRoutingRules(rules, vsPort) {
    const tbody = document.getElementById('rules-table').querySelector('tbody');
    
    // Save current scroll position
    const scrollPos = tbody.scrollTop;
    
    // Clear table with fade-out animation
    tbody.style.opacity = '0.5';
    
    setTimeout(() => {
        tbody.innerHTML = '';
        
        if (!rules || rules.length === 0) {
            const tr = document.createElement('tr');
            tr.innerHTML = '<td colspan="5" class="text-center">No content routing rules defined for this virtual service.</td>';
            tbody.appendChild(tr);
            tbody.style.opacity = '1';
            return;
        }
        
        rules.forEach((rule, index) => {
            const tr = document.createElement('tr');
            
            tr.innerHTML = `
                <td>${index + 1}</td>
                <td>${rule.key}</td>
                <td>${rule.value}</td>
                <td>${rule.serverName}</td>
                <td>
                    <button class="btn btn-sm btn-danger delete-rule-btn" data-vs="${vsPort}" data-index="${index}">
                        <i class="fas fa-trash me-1"></i>Delete
                    </button>
                </td>
            `;
            
            // Add with fade-in animation
            tr.style.opacity = '0';
            tbody.appendChild(tr);
            setTimeout(() => {
                tr.style.opacity = '1';
            }, 50);
        });
        
        // Restore table opacity
        setTimeout(() => {
            tbody.style.opacity = '1';
            
            // Restore scroll position
            tbody.scrollTop = scrollPos;
            
            // Add event listeners for delete button
            document.querySelectorAll('.delete-rule-btn').forEach(btn => {
                btn.addEventListener('click', () => deleteContentRoutingRule(btn.dataset.vs, btn.dataset.index));
            });
        }, 100);
    }, 200);
}

// Add a server to the VS form
function addServerToForm() {
    const serverList = document.getElementById('server-list');
    const serverTemplate = serverList.children[0].cloneNode(true);
    
    // Clear input values
    serverTemplate.querySelectorAll('input').forEach(input => {
        input.value = input.classList.contains('server-weight') ? '1' : '';
    });
    
    // Add remove button if it doesn't exist
    if (!serverTemplate.querySelector('.remove-server-btn')) {
        const removeBtn = document.createElement('button');
        removeBtn.type = 'button';
        removeBtn.className = 'btn btn-sm btn-danger remove-server-btn';
        removeBtn.innerHTML = '<i class="fas fa-times"></i>';
        removeBtn.addEventListener('click', function() {
            const serverItem = this.closest('.server-item');
            // Add fade-out animation
            serverItem.style.opacity = '0';
            serverItem.style.transform = 'translateX(10px)';
            setTimeout(() => {
                serverItem.remove();
            }, 300);
        });
        
        serverTemplate.appendChild(removeBtn);
    }
    
    // Add new server with animation
    serverTemplate.style.opacity = '0';
    serverTemplate.style.transform = 'translateY(10px)';
    serverList.appendChild(serverTemplate);
    
    // Trigger animation
    setTimeout(() => {
        serverTemplate.style.opacity = '1';
        serverTemplate.style.transform = 'translateY(0)';
    }, 10);
    
    // Add event listener to the new remove button
    serverTemplate.querySelector('.remove-server-btn').addEventListener('click', function() {
        const serverItem = this.closest('.server-item');
        // Add fade-out animation
        serverItem.style.opacity = '0';
        serverItem.style.transform = 'translateX(10px)';
        setTimeout(() => {
            serverItem.remove();
        }, 300);
    });
    
    // Focus the first input
    setTimeout(() => {
        serverTemplate.querySelector('.server-name').focus();
    }, 300);
}

// Save a new virtual service
async function saveVirtualService() {
    // Validate form
    const form = document.getElementById('add-vs-form');
    if (!validateForm(form)) {
        return;
    }
    
    // Show a loading spinner in the save button
    const saveBtn = document.getElementById('save-vs-btn');
    const originalBtnHtml = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Saving...';
    saveBtn.disabled = true;
    
    // Get form values
    const port = document.getElementById('vs-port').value;
    const algorithm = document.getElementById('vs-algorithm').value;
    const rateLimit = document.getElementById('vs-rate-limit').value;
    const statusCode = document.getElementById('vs-status-code').value;
    const message = document.getElementById('vs-message').value;
    
    // Get server list
    const serverItems = document.querySelectorAll('.server-item');
    const serverList = [];
    
    for (const item of serverItems) {
        const name = item.querySelector('.server-name').value;
        const url = item.querySelector('.server-url').value;
        const weight = parseInt(item.querySelector('.server-weight').value);
        
        if (!name || !url) {
            showAlert('Please fill in all server details', 'danger');
            
            // Restore button state
            saveBtn.innerHTML = originalBtnHtml;
            saveBtn.disabled = false;
            
            return;
        }
        
        serverList.push({
            name,
            url,
            weight
        });
    }
    
    // Create payload
    const payload = {
        port: parseInt(port),
        algorithm,
        serverList: serverList,
        rate_limit: parseInt(rateLimit) || 0,
        status_code: parseInt(statusCode),
        message
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': AUTH_CREDENTIALS
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
        }
        
        // Close modal and reload data
        document.querySelector('#addVSModal .btn-close').click();
        
        // Show success message
        showAlert('Virtual service created successfully', 'success');
        
        // Reset form
        form.reset();
        const serverList = document.getElementById('server-list');
        while (serverList.children.length > 1) {
            serverList.removeChild(serverList.lastChild);
        }
        
        // Reload data
        await loadDashboardData();
    } catch (error) {
        console.error('Error creating virtual service:', error);
        showAlert(`Failed to create virtual service: ${error.message}`, 'danger');
    } finally {
        // Restore button state
        saveBtn.innerHTML = originalBtnHtml;
        saveBtn.disabled = false;
    }
}

// Validate form
function validateForm(form) {
    let isValid = true;
    
    // Check each required input
    form.querySelectorAll('[required]').forEach(input => {
        if (!input.value.trim()) {
            input.classList.add('is-invalid');
            isValid = false;
            
            // Add event listener to remove invalid class on input
            input.addEventListener('input', function() {
                if (this.value.trim()) {
                    this.classList.remove('is-invalid');
                }
            }, { once: true });
        }
    });
    
    if (!isValid) {
        showAlert('Please fill in all required fields', 'danger');
    }
    
    return isValid;
}

// View a virtual service in a detailed modal
function viewVirtualService(port) {
    const vs = virtualServices.find(v => v.port == port);
    if (!vs) {
        showAlert('Virtual service not found', 'danger');
        return;
    }
    
    // Create a modal to show the details
    const modalId = 'vsDetailsModal';
    let modal = document.getElementById(modalId);
    
    // Remove existing modal if it exists
    if (modal) {
        document.body.removeChild(modal);
    }
    
    // Create new modal
    modal = document.createElement('div');
    modal.className = 'modal fade';
    modal.id = modalId;
    modal.tabIndex = '-1';
    modal.setAttribute('aria-labelledby', `${modalId}Label`);
    modal.setAttribute('aria-hidden', 'true');
    
    // Format algorithm name for display
    const algorithmDisplay = vs.algorithm.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    
    // Create server rows HTML
    let serverRowsHtml = '';
    vs.serverList.forEach((server, index) => {
        const healthStatus = server.health ? 
            '<span class="badge bg-success">Healthy</span>' : 
            '<span class="badge bg-danger">Unhealthy</span>';
            
        serverRowsHtml += `
            <tr>
                <td>${index + 1}</td>
                <td>${server.name}</td>
                <td>${server.url}</td>
                <td>${server.weight}</td>
                <td>${healthStatus}</td>
                <td>${server.active ? 'Yes' : 'No'}</td>
            </tr>
        `;
    });
    
    // Check if VS has SSL certificate
    const hasSSL = certificates[vs.port] ? true : false;
    const sslStatus = hasSSL ? 
        '<span class="badge bg-success">Enabled</span>' : 
        '<span class="badge bg-secondary">Disabled</span>';
    
    // Build modal content
    modal.innerHTML = `
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header bg-primary text-white">
                    <h5 class="modal-title" id="${modalId}Label">
                        <i class="fas fa-server me-2"></i>Virtual Service Details: Port ${vs.port}
                    </h5>
                    <button type="button" class="btn-close btn-close-white" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="row mb-4">
                        <div class="col-md-6">
                            <h6 class="fw-bold">General Information</h6>
                            <table class="table table-sm">
                                <tr>
                                    <th>Port:</th>
                                    <td>${vs.port}</td>
                                </tr>
                                <tr>
                                    <th>Algorithm:</th>
                                    <td>${algorithmDisplay}</td>
                                </tr>
                                <tr>
                                    <th>SSL:</th>
                                    <td>${sslStatus}</td>
                                </tr>
                            </table>
                        </div>
                        <div class="col-md-6">
                            <h6 class="fw-bold">Rate Limiting</h6>
                            <table class="table table-sm">
                                <tr>
                                    <th>Rate Limit:</th>
                                    <td>${vs.rate_limit || 'Not set'}</td>
                                </tr>
                                <tr>
                                    <th>Status Code:</th>
                                    <td>${vs.status_code || '-'}</td>
                                </tr>
                                <tr>
                                    <th>Message:</th>
                                    <td>${vs.message || '-'}</td>
                                </tr>
                            </table>
                        </div>
                    </div>
                    
                    <h6 class="fw-bold">Server List</h6>
                    <div class="table-responsive">
                        <table class="table table-striped table-hover table-sm">
                            <thead>
                                <tr>
                                    <th>#</th>
                                    <th>Name</th>
                                    <th>URL</th>
                                    <th>Weight</th>
                                    <th>Health</th>
                                    <th>Active</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${serverRowsHtml}
                            </tbody>
                        </table>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                    <button type="button" class="btn btn-warning edit-vs-btn-modal" data-port="${vs.port}">
                        <i class="fas fa-edit me-2"></i>Edit
                    </button>
                </div>
            </div>
        </div>
    `;
    
    // Add modal to body
    document.body.appendChild(modal);
    
    // Initialize modal and show it
    const modalObj = new bootstrap.Modal(modal);
    modalObj.show();
    
    // Add event listener for edit button in modal
    modal.querySelector('.edit-vs-btn-modal').addEventListener('click', () => {
        modalObj.hide();
        editVirtualService(vs.port);
    });
}

// Edit a virtual service
function editVirtualService(port) {
    const vs = virtualServices.find(v => v.port == port);
    if (!vs) {
        showAlert('Virtual service not found', 'danger');
        return;
    }
    
    // Get the form elements
    const form = document.getElementById('add-vs-form');
    const modalTitle = document.getElementById('addVSModalLabel');
    const saveBtn = document.getElementById('save-vs-btn');
    
    // Update modal title and save button to reflect edit mode
    modalTitle.innerHTML = `<i class="fas fa-edit me-2"></i>Edit Virtual Service (Port ${vs.port})`;
    saveBtn.innerHTML = '<i class="fas fa-save me-2"></i>Update';
    
    // Store the original port for reference in the update process
    form.dataset.originalPort = vs.port;
    
    // Fill the form with the VS details
    document.getElementById('vs-port').value = vs.port;
    document.getElementById('vs-algorithm').value = vs.algorithm;
    document.getElementById('vs-rate-limit').value = vs.rate_limit || 0;
    document.getElementById('vs-status-code').value = vs.status_code || 429;
    document.getElementById('vs-message').value = vs.message || 'Rate limit exceeded';
    
    // Clear existing server list except for the first template item
    const serverList = document.getElementById('server-list');
    while (serverList.children.length > 1) {
        serverList.removeChild(serverList.lastChild);
    }
    
    // Fill in the first server item
    if (vs.serverList && vs.serverList.length > 0) {
        const firstServer = vs.serverList[0];
        const firstServerItem = serverList.children[0];
        
        firstServerItem.querySelector('.server-name').value = firstServer.name;
        firstServerItem.querySelector('.server-url').value = firstServer.url;
        firstServerItem.querySelector('.server-weight').value = firstServer.weight || 1;
        
        // Add the rest of the servers
        for (let i = 1; i < vs.serverList.length; i++) {
            addServerToForm();
            const serverItem = serverList.children[i];
            const server = vs.serverList[i];
            
            serverItem.querySelector('.server-name').value = server.name;
            serverItem.querySelector('.server-url').value = server.url;
            serverItem.querySelector('.server-weight').value = server.weight || 1;
        }
    }
    
    // Change the save button click handler to update instead of create
    saveBtn.onclick = updateVirtualService;
    
    // Show the modal
    const addVSModal = new bootstrap.Modal(document.getElementById('addVSModal'));
    addVSModal.show();
}

// Update an existing virtual service
async function updateVirtualService() {
    // Validate form
    const form = document.getElementById('add-vs-form');
    if (!validateForm(form)) {
        return;
    }
    
    // Get the original port from the form dataset
    const originalPort = form.dataset.originalPort;
    
    // Show a loading spinner in the save button
    const saveBtn = document.getElementById('save-vs-btn');
    const originalBtnHtml = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Updating...';
    saveBtn.disabled = true;
    
    // Get form values
    const port = document.getElementById('vs-port').value;
    const algorithm = document.getElementById('vs-algorithm').value;
    const rateLimit = document.getElementById('vs-rate-limit').value;
    const statusCode = document.getElementById('vs-status-code').value;
    const message = document.getElementById('vs-message').value;
    
    // Get server list
    const serverItems = document.querySelectorAll('.server-item');
    const serverList = [];
    
    for (const item of serverItems) {
        const name = item.querySelector('.server-name').value;
        const url = item.querySelector('.server-url').value;
        const weight = parseInt(item.querySelector('.server-weight').value);
        
        if (!name || !url) {
            showAlert('Please fill in all server details', 'danger');
            
            // Restore button state
            saveBtn.innerHTML = originalBtnHtml;
            saveBtn.disabled = false;
            
            return;
        }
        
        serverList.push({
            name,
            url,
            weight
        });
    }
    
    // Create payload
    const payload = {
        port: parseInt(port),
        algorithm,
        serverList: serverList,
        rate_limit: parseInt(rateLimit) || 0,
        status_code: parseInt(statusCode),
        message
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/${originalPort}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': AUTH_CREDENTIALS
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
        }
        
        // Close modal
        document.querySelector('#addVSModal .btn-close').click();
        
        // Show success message
        showAlert('Virtual service updated successfully', 'success');
        
        // Reset form and restore original save handler
        form.reset();
        delete form.dataset.originalPort;
        document.getElementById('addVSModalLabel').innerHTML = '<i class="fas fa-server me-2"></i>Add Virtual Service';
        saveBtn.innerHTML = '<i class="fas fa-save me-2"></i>Save';
        saveBtn.onclick = saveVirtualService;
        
        const serverList = document.getElementById('server-list');
        while (serverList.children.length > 1) {
            serverList.removeChild(serverList.lastChild);
        }
        
        // Reload data
        await loadDashboardData();
    } catch (error) {
        console.error('Error updating virtual service:', error);
        showAlert(`Failed to update virtual service: ${error.message}`, 'danger');
    } finally {
        // Restore button state
        saveBtn.innerHTML = originalBtnHtml;
        saveBtn.disabled = false;
        saveBtn.onclick = saveVirtualService;
    }
}

// Delete a virtual service
async function deleteVirtualService(port) {
    // Create a custom confirmation modal instead of using browser confirm
    const confirmDelete = await showConfirmDialog(
        'Delete Virtual Service',
        `Are you sure you want to delete the virtual service on port ${port}?`,
        'This action cannot be undone.'
    );
    
    if (!confirmDelete) {
        return;
    }
    
    // Show loading overlay during deletion
    showLoadingOverlay();
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/${port}`, {
            method: 'DELETE',
            headers: {
                'Authorization': AUTH_CREDENTIALS
            }
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
        }
        
        showAlert('Virtual service deleted successfully', 'success');
        
        // Remove the virtual service from the local array to reflect changes immediately
        const index = virtualServices.findIndex(vs => vs.port == port);
        if (index !== -1) {
            virtualServices.splice(index, 1);
        }
        
        // Reload data to update UI
        loadDashboardData();
    } catch (error) {
        console.error('Error deleting virtual service:', error);
        showAlert(`Failed to delete virtual service: ${error.message}`, 'danger');
    } finally {
        // Hide loading overlay
        hideLoadingOverlay();
    }
}

// Generate an SSL certificate
async function generateCertificate() {
    // Validate form
    const form = document.getElementById('generate-cert-form');
    if (!validateForm(form)) {
        return;
    }
    
    // Show a loading spinner in the generate button
    const generateBtn = document.getElementById('generate-cert-btn');
    const originalBtnHtml = generateBtn.innerHTML;
    generateBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Generating...';
    generateBtn.disabled = true;
    
    // Get form values
    const port = document.getElementById('cert-port').value;
    const commonName = document.getElementById('cert-common-name').value;
    const days = document.getElementById('cert-days').value;
    
    // Create payload
    const payload = {
        port: parseInt(port),
        commonName,
        days: parseInt(days)
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/certificates/generate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': AUTH_CREDENTIALS
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Close modal
        document.querySelector('#generateCertModal .btn-close').click();
        
        // Show success message
        showAlert('Certificate generated successfully', 'success');
        
        // Reset form
        form.reset();
        
        // Reload data
        await loadDashboardData();
    } catch (error) {
        console.error('Error generating certificate:', error);
        showAlert(`Failed to generate certificate: ${error.message}`, 'danger');
    } finally {
        // Restore button state
        generateBtn.innerHTML = originalBtnHtml;
        generateBtn.disabled = false;
    }
}

// Renew an SSL certificate
async function renewCertificate(port) {
    // Create a custom confirmation modal
    const confirmRenew = await showConfirmDialog(
        'Renew SSL Certificate',
        `Are you sure you want to renew the certificate for port ${port}?`,
        'The existing certificate will be replaced.'
    );
    
    if (!confirmRenew) {
        return;
    }
    
    // Show loading overlay during renewal
    showLoadingOverlay();
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/certificates/renew/${port}`, {
            method: 'POST',
            headers: {
                'Authorization': AUTH_CREDENTIALS
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        showAlert('Certificate renewed successfully', 'success');
        
        // Reload data
        await loadDashboardData();
    } catch (error) {
        console.error('Error renewing certificate:', error);
        showAlert(`Failed to renew certificate: ${error.message}`, 'danger');
    } finally {
        // Hide loading overlay
        hideLoadingOverlay();
    }
}

// Block an IP address
async function blockIP() {
    // Validate form
    const form = document.getElementById('add-ip-form');
    if (!validateForm(form)) {
        return;
    }
    
    // Show a loading spinner in the block button
    const blockBtn = document.getElementById('block-ip-btn');
    const originalBtnHtml = blockBtn.innerHTML;
    blockBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Blocking...';
    blockBtn.disabled = true;
    
    // Get form values
    const ipAddress = document.getElementById('ip-address').value;
    
    // Create payload
    const payload = {
        rule: 'block',
        ip: ipAddress
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/ip-rules`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': AUTH_CREDENTIALS
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Close modal
        document.querySelector('#addIpRuleModal .btn-close').click();
        
        // Show success message
        showAlert('IP address blocked successfully', 'success');
        
        // Reset form
        form.reset();
        
        // Add to our local list for display
        blacklistedIPs.push({
            ip: ipAddress,
            timestamp: new Date().toLocaleString()
        });
        
        // Update display
        displayBlacklistedIPs();
    } catch (error) {
        console.error('Error blocking IP address:', error);
        showAlert(`Failed to block IP address: ${error.message}`, 'danger');
    } finally {
        // Restore button state
        blockBtn.innerHTML = originalBtnHtml;
        blockBtn.disabled = false;
    }
}

// Open the edit rate limit modal with prefilled values
function editRateLimit(port) {
    const vs = virtualServices.find(v => v.port == port);
    if (!vs) {
        showAlert('Virtual service not found', 'danger');
        return;
    }
    
    // Prefill form with animation
    document.getElementById('rate-limit-port').value = vs.port;
    
    const valueInput = document.getElementById('rate-limit-value');
    const statusInput = document.getElementById('rate-limit-status-code');
    const messageInput = document.getElementById('rate-limit-message');
    
    // Highlight the changes with animation
    valueInput.style.transition = 'background-color 0.3s ease';
    statusInput.style.transition = 'background-color 0.3s ease';
    messageInput.style.transition = 'background-color 0.3s ease';
    
    valueInput.style.backgroundColor = '#f8f9fa';
    statusInput.style.backgroundColor = '#f8f9fa';
    messageInput.style.backgroundColor = '#f8f9fa';
    
    valueInput.value = vs.rate_limit || 0;
    statusInput.value = vs.status_code || 429;
    messageInput.value = vs.message || 'Rate limit exceeded';
    
    setTimeout(() => {
        valueInput.style.backgroundColor = '';
        statusInput.style.backgroundColor = '';
        messageInput.style.backgroundColor = '';
    }, 500);
    
    // Show modal
    new bootstrap.Modal(document.getElementById('updateRateLimitModal')).show();
}

// Update a rate limit
async function updateRateLimit() {
    // Validate form
    const form = document.getElementById('update-rate-limit-form');
    if (!validateForm(form)) {
        return;
    }
    
    // Show a loading spinner in the update button
    const updateBtn = document.getElementById('update-rate-limit-btn');
    const originalBtnHtml = updateBtn.innerHTML;
    updateBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Updating...';
    updateBtn.disabled = true;
    
    // Get form values
    const port = document.getElementById('rate-limit-port').value;
    const rateLimit = document.getElementById('rate-limit-value').value;
    const statusCode = document.getElementById('rate-limit-status-code').value;
    const message = document.getElementById('rate-limit-message').value;
    
    // Create payload
    const payload = {
        port: parseInt(port),
        rate_limit: parseInt(rateLimit),
        status_code: parseInt(statusCode),
        message
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/rate-limits`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': AUTH_CREDENTIALS
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Close modal
        document.querySelector('#updateRateLimitModal .btn-close').click();
        
        // Show success message
        showAlert('Rate limit updated successfully', 'success');
        
        // Reload data
        await loadDashboardData();
    } catch (error) {
        console.error('Error updating rate limit:', error);
        showAlert(`Failed to update rate limit: ${error.message}`, 'danger');
    } finally {
        // Restore button state
        updateBtn.innerHTML = originalBtnHtml;
        updateBtn.disabled = false;
    }
}

// Save a content routing rule
async function saveContentRoutingRule() {
    // Validate form
    const form = document.getElementById('add-rule-form');
    if (!validateForm(form)) {
        return;
    }
    
    // Show a loading spinner in the save button
    const saveBtn = document.getElementById('save-rule-btn');
    const originalBtnHtml = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Saving...';
    saveBtn.disabled = true;
    
    // Get form values
    const vsPort = document.getElementById('rule-vs-port').value;
    const headerKey = document.getElementById('rule-header-key').value;
    const headerValue = document.getElementById('rule-header-value').value;
    const serverName = document.getElementById('rule-server-name').value;
    
    // Create payload
    const payload = {
        key: headerKey,
        value: headerValue,
        serverName
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/${vsPort}/rules`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': AUTH_CREDENTIALS
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Close modal
        document.querySelector('#addRuleModal .btn-close').click();
        
        // Show success message
        showAlert('Content routing rule added successfully', 'success');
        
        // Reset form
        form.reset();
        
        // Reload rules
        loadContentRoutingRules();
    } catch (error) {
        console.error('Error adding content routing rule:', error);
        showAlert(`Failed to add content routing rule: ${error.message}`, 'danger');
    } finally {
        // Restore button state
        saveBtn.innerHTML = originalBtnHtml;
        saveBtn.disabled = false;
    }
}

// Delete a content routing rule
async function deleteContentRoutingRule(vsPort, ruleIndex) {
    // Create a custom confirmation modal
    const confirmDelete = await showConfirmDialog(
        'Delete Routing Rule',
        'Are you sure you want to delete this routing rule?',
        'This action cannot be undone.'
    );
    
    if (!confirmDelete) {
        return;
    }
    
    // Show loading overlay during deletion
    showLoadingOverlay();
    
    try {
        const response = await fetch(`${API_BASE_URL}/access/vs/${vsPort}/rules/${ruleIndex}`, {
            method: 'DELETE',
            headers: {
                'Authorization': AUTH_CREDENTIALS
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        showAlert('Content routing rule deleted successfully', 'success');
        
        // Reload rules
        loadContentRoutingRules();
    } catch (error) {
        console.error('Error deleting content routing rule:', error);
        showAlert(`Failed to delete content routing rule: ${error.message}`, 'danger');
    } finally {
        // Hide loading overlay
        hideLoadingOverlay();
    }
}

// Show confirmation dialog
function showConfirmDialog(title, message, details) {
    return new Promise((resolve) => {
        // Create modal element
        const modal = document.createElement('div');
        modal.className = 'modal fade';
        modal.id = 'confirmModal';
        modal.tabIndex = '-1';
        modal.setAttribute('aria-labelledby', 'confirmModalLabel');
        modal.setAttribute('aria-hidden', 'true');
        
        modal.innerHTML = `
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title" id="confirmModalLabel"><i class="fas fa-exclamation-triangle text-warning me-2"></i>${title}</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        <p>${message}</p>
                        <p class="text-muted small">${details}</p>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-danger" id="confirm-yes">Yes, proceed</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Initialize the modal
        const modalObj = new bootstrap.Modal(modal);
        modalObj.show();
        
        // Add event listeners
        modal.querySelector('#confirm-yes').addEventListener('click', () => {
            modalObj.hide();
            resolve(true);
        });
        
        modal.addEventListener('hidden.bs.modal', () => {
            document.body.removeChild(modal);
            resolve(false);
        });
    });
}

// Helper function to show alerts
function showAlert(message, type = 'info') {
    // Check if an alert container exists, create one if not
    let alertContainer = document.getElementById('alert-container');
    if (!alertContainer) {
        alertContainer = document.createElement('div');
        alertContainer.id = 'alert-container';
        alertContainer.style.position = 'fixed';
        alertContainer.style.top = '20px';
        alertContainer.style.right = '20px';
        alertContainer.style.zIndex = '1050';
        document.body.appendChild(alertContainer);
    }
    
    // Create alert element
    const alert = document.createElement('div');
    alert.className = `alert alert-${type} alert-dismissible fade show`;
    
    // Add icon based on alert type
    let icon = 'info-circle';
    if (type === 'success') icon = 'check-circle';
    if (type === 'danger') icon = 'exclamation-circle';
    if (type === 'warning') icon = 'exclamation-triangle';
    
    alert.innerHTML = `
        <i class="fas fa-${icon} me-2"></i>${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;
    
    // Add alert to container
    alertContainer.appendChild(alert);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        if (alert && alert.parentNode === alertContainer) {
            alert.classList.remove('show');
            setTimeout(() => {
                if (alert.parentNode === alertContainer) {
                    alertContainer.removeChild(alert);
                }
            }, 300);
        }
    }, 5000);
}