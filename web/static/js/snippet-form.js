(() => {
  const form = document.querySelector("[data-snippet-form]");
  if (!form) {
    return;
  }

  const errorBanner = form.querySelector("[data-form-error]");
  const submitButton = form.querySelector('button[type="submit"]');

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    clearErrors();

    const titleInput = form.elements.namedItem("title");
    const descriptionInput = form.elements.namedItem("description");
    const languageInput = form.elements.namedItem("language");
    const contentInput = form.elements.namedItem("content");
    const publicInput = form.elements.namedItem("is_public");

    const payload = {
      title: titleInput ? titleInput.value.trim() : "",
      description: descriptionInput ? descriptionInput.value.trim() : "",
      language: languageInput ? languageInput.value : "",
      content: contentInput ? contentInput.value : "",
      is_public: publicInput ? publicInput.checked : false,
    };

    const snippetId = form.dataset.snippetId;
    const isEditMode = form.dataset.mode === "edit";
    const url = isEditMode ? `/api/snippets/${snippetId}` : "/api/snippets";
    const method = isEditMode ? "PUT" : "POST";

    toggleSubmitState(true);

    try {
      const response = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (response.ok) {
        const body = await response.json();
        window.location.href = `/snippets/${body.data.id}`;
        return;
      }

      const body = await response.json().catch(() => ({}));
      if (response.status === 422 && body.fields) {
        renderFieldErrors(body.fields);
        showBanner(body.message || "Please fix the highlighted fields.");
        return;
      }

      showBanner(body.message || "The snippet could not be saved. Please try again.");
    } catch (error) {
      showBanner("The request failed. Please check your connection and try again.");
    } finally {
      toggleSubmitState(false);
    }
  });

  function clearErrors() {
    errorBanner.hidden = true;
    errorBanner.textContent = "";
    form.querySelectorAll("[data-field-error]").forEach((element) => {
      element.textContent = "";
    });
  }

  function renderFieldErrors(fieldErrors) {
    Object.entries(fieldErrors).forEach(([fieldName, message]) => {
      const target = form.querySelector(`[data-field-error="${fieldName}"]`);
      if (target) {
        target.textContent = message;
      }
    });
  }

  function showBanner(message) {
    errorBanner.hidden = false;
    errorBanner.textContent = message;
  }

  function toggleSubmitState(isLoading) {
    if (!submitButton) {
      return;
    }

    submitButton.disabled = isLoading;
    submitButton.textContent = isLoading ? "Saving..." : form.dataset.mode === "edit" ? "Save changes" : "Create snippet";
  }
})();
