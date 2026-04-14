// High-contrast color palette optimized for bright outdoor environments (farms)
const FIELD_PROD_COLORS = [
    { bg: 'rgba(13, 148, 136, 0.85)', border: 'rgba(13, 148, 136, 1)' },   // Teal 600
    { bg: 'rgba(234, 88, 12, 0.85)', border: 'rgba(234, 88, 12, 1)' },    // Orange 600
    { bg: 'rgba(37, 99, 235, 0.85)', border: 'rgba(37, 99, 235, 1)' },    // Blue 600
    { bg: 'rgba(124, 58, 237, 0.85)', border: 'rgba(124, 58, 237, 1)' },  // Violet 600
    { bg: 'rgba(217, 119, 6, 0.85)', border: 'rgba(217, 119, 6, 1)' }     // Amber 600
];

function getColorArray(dataLength, type) {
    const colors = [];
    for (let i = 0; i < dataLength; i++) {
        colors.push(FIELD_PROD_COLORS[i % FIELD_PROD_COLORS.length][type]);
    }
    return colors;
}

export function setupFieldProdChart(elementId, title, labels, data) {
    new Chart(document.getElementById(elementId), {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [
                {
                    data: data,
                    backgroundColor: getColorArray(data.length, 'bg'),
                    borderColor: getColorArray(data.length, 'border'),
                    borderWidth: 2
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true
                }
            },
            plugins: {
                legend: {
                    display: false
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
