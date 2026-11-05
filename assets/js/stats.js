import { formatWeightElements } from "/public/assets/js/weight.js";

document.body.addEventListener('htmx:afterSwap', function(evt) {
  if (evt.detail.elt) {
    formatWeightElements(evt.detail.elt);
  }
});
