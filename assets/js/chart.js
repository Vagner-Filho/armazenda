// High-contrast color palette optimized for bright outdoor environments (farms)
const FIELD_PROD_COLORS = [
    { bg: 'rgba(13, 148, 136, 0.85)', border: 'rgba(13, 148, 136, 1)' },   // Teal 600
    { bg: 'rgba(234, 88, 12, 0.85)', border: 'rgba(234, 88, 12, 1)' },    // Orange 600
    { bg: 'rgba(37, 99, 235, 0.85)', border: 'rgba(37, 99, 235, 1)' },    // Blue 600
    { bg: 'rgba(124, 58, 237, 0.85)', border: 'rgba(124, 58, 237, 1)' },  // Violet 600
    { bg: 'rgba(217, 119, 6, 0.85)', border: 'rgba(217, 119, 6, 1)' }     // Amber 600
];

// Representative colors for the well-known products; any other product falls
// back to the generic palette by dataset order.
const PRODUCT_COLORS = {
    'milho': { bg: 'rgba(245, 158, 11, 0.85)', border: 'rgba(245, 158, 11, 1)' },   // Amber (corn)
    'soja':  { bg: 'rgba(34, 197, 94, 0.85)',  border: 'rgba(34, 197, 94, 1)' }     // Green (soy)
};

function getProductColor(productName, fallbackIndex) {
    const normalized = (productName || '').trim().toLowerCase();
    if (PRODUCT_COLORS[normalized]) {
        return PRODUCT_COLORS[normalized];
    }
    return FIELD_PROD_COLORS[fallbackIndex % FIELD_PROD_COLORS.length];
}

/**
 * Stacked bar chart: labels are fields and each dataset is a product, so
 * every field's bar is stacked by product. Colors are assigned per dataset
 * index so a product keeps the same color across all charts on the page.
 * @param {string} elementId - canvas element id
 * @param {string} title - chart title
 * @param {string[]} labels - field names
 * @param {{product: string, values: number[]}[]} datasets - one series per product
 */
export function setupProductStackedChart(elementId, title, labels, datasets) {
    new Chart(document.getElementById(elementId), {
        type: 'bar',
        data: {
            labels: labels,
            datasets: datasets.map((dataset, index) => {
                const color = getProductColor(dataset.product, index);
                return {
                    label: dataset.product,
                    data: dataset.values,
                    backgroundColor: color.bg,
                    borderColor: color.border,
                    borderWidth: 2
                };
            })
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    stacked: true
                },
                y: {
                    stacked: true,
                    beginAtZero: true
                }
            },
            plugins: {
                legend: {
                    display: true,
                    labels: {
                        color: '#ffffff99'
                    }
                },
                title: {
                    display: true,
                    text: title,
                    color: '#ffffff99',
                    font: {
                        size: 13,
                        weight: 'normal'
                    },
                    padding: {
                        top: 5,
                        bottom: 15
                    }
                }
            },
        }
    });
}
