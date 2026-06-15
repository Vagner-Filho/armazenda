// Generic filter-modal controller using event delegation on document.body.
//
// Why delegation: parts of the romaneio page (entry-content / departure-content)
// are swapped via HTMX at runtime, so any dialog or button bound at page-load
// time would lose its listeners after a toggle. Listening at body level
// catches every click regardless of when the element was inserted.
//
// Conventions:
//   - The modal: <dialog data-filter-modal data-filter-path="/post/path">
//   - Open from anywhere: <button data-opens="<dialog-id>">
//   - Close from inside the modal: <button data-closes>
//   - Backdrop click on the dialog itself also closes
//   - On successful htmx response to data-filter-path, the matching dialog closes
//   - Chip remove: <button data-remove-chip="field" data-remove-chip-form="form-id">

function handleClick(e) {
  // Open button
  const openBtn = e.target.closest('[data-opens]');
  if (openBtn) {
    const id = openBtn.dataset.opens;
    const dialog = document.getElementById(id);
    console.log('[filterModal] open click', { id, dialog, hasShowModal: typeof dialog?.showModal });
    if (dialog && typeof dialog.showModal === 'function') {
      dialog.showModal();
    }
    return;
  }

  // Close button (anywhere inside a filter dialog)
  const closeBtn = e.target.closest('[data-closes]');
  if (closeBtn) {
    closeBtn.closest('dialog[data-filter-modal]')?.close();
    return;
  }

  // Backdrop click — fires when the click target IS the dialog element itself.
  if (e.target.matches?.('dialog[data-filter-modal]')) {
    const dialog = e.target;
    const rect = dialog.getBoundingClientRect();
    const inDialog =
      e.clientX >= rect.left && e.clientX <= rect.right &&
      e.clientY >= rect.top && e.clientY <= rect.bottom;
    if (!inDialog) dialog.close();
    return;
  }

  // Chip removal
  const chipBtn = e.target.closest('[data-remove-chip][data-remove-chip-form]');
  if (chipBtn) {
    const key = chipBtn.dataset.removeChip;
    const formId = chipBtn.dataset.removeChipForm;
    const form = document.getElementById(formId);
    if (!form) return;
    const input = form.querySelector(`[name="${key}"]`);
    if (input) input.value = '';
    if (window.htmx) window.htmx.trigger(form, 'submit');
  }
}

function handleHtmxAfter(e) {
  if (!e.detail?.successful) return;
  const path = e.detail?.requestConfig?.path;
  if (!path) return;
  document.querySelectorAll(`dialog[data-filter-modal][data-filter-path="${path}"]`).forEach((dialog) => {
    if (dialog.open) dialog.close();
  });
}

function init() {
  console.log('[filterModal] init — body click listener attached');
  document.body.addEventListener('click', handleClick);
  document.body.addEventListener('htmx:afterRequest', handleHtmxAfter);
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}
