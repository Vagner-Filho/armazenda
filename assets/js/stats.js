document.addEventListener("DOMContentLoaded", function() {
  const statCards = document.querySelectorAll("[data-stat-type]");

  const statStyles = {
    "top_supplier": { icon: "mdi:truck-check", color: "#4CAF50" },
    "top_buyer": { icon: "mdi:cart-check", color: "#2196F3" },
    "most_frequent_supplier": { icon: "mdi:truck-fast", color: "#FF9800" },
    "best_quality_supplier": { icon: "mdi:star-check", color: "#8BC34A" },
    "worst_quality_supplier": { icon: "mdi:alert-circle", color: "#F44336" }
  };

  statCards.forEach(card => {
    const type = card.dataset.statType;
    const style = statStyles[type];

    if (style) {
      const iconElement = document.createElement("iconify-icon");
      iconElement.setAttribute("icon", style.icon);
      iconElement.setAttribute("width", "32");
      iconElement.setAttribute("height", "32");
      iconElement.style.color = style.color;

      card.querySelector("h3").insertAdjacentElement("beforebegin", iconElement);
    }
  });
});
