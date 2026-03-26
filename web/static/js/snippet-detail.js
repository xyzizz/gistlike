(() => {
  const deleteButton = document.querySelector("[data-delete-snippet]");
  if (!deleteButton) {
    return;
  }

  deleteButton.addEventListener("click", async () => {
    const snippetId = deleteButton.dataset.snippetId;
    if (!snippetId) {
      return;
    }

    const confirmed = window.confirm("Delete this snippet permanently?");
    if (!confirmed) {
      return;
    }

    deleteButton.disabled = true;
    const originalText = deleteButton.textContent;
    deleteButton.textContent = "Deleting...";

    try {
      const response = await fetch(`/api/snippets/${snippetId}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
        },
      });

      if (response.status === 204) {
        window.location.href = "/";
        return;
      }

      const body = await response.json().catch(() => ({}));
      window.alert(body.message || "The snippet could not be deleted.");
    } catch (error) {
      window.alert("The delete request failed. Please try again.");
    } finally {
      deleteButton.disabled = false;
      deleteButton.textContent = originalText;
    }
  });
})();
