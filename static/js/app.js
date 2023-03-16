document.querySelectorAll(".nav-link").forEach((link) => {
  if (link.href === window.location.href) {
    link.classList.add("active");
    link.setAttribute("aria-current", "page");
  }
});

(() => {
  "use strict";

  // Fetch all the forms we want to apply custom Bootstrap validation styles to
  // Loop over them and prevent submission
  const forms = document.querySelectorAll(".needs-validation");
  Array.from(forms).forEach((form) => {
    form.addEventListener(
      "submit",
      (event) => {
        if (!form.checkValidity()) {
          event.preventDefault();
          event.stopPropagation();
        }

        form.classList.add("was-validated");
      },
      false
    );
  });
})();

const Prompt = () => {
  const toast = (c) => {
    const {
      title = "",
      icon = "success",
      position = "top-end",
      timer = 4000,
      showConfirmButton = false,
      confirmButtonText = "Got it",
    } = c;

    const Toast = Swal.mixin({
      toast: true,
      title,
      icon,
      position,
      showConfirmButton,
      confirmButtonText,
      timer,
      timerProgressBar: true,
      didOpen: (toast) => {
        toast.addEventListener("mouseenter", Swal.stopTimer);
        toast.addEventListener("mouseleave", Swal.resumeTimer);
      },
    });

    Toast.fire({});
  };

  const alertModal = (c) => {
    const { icon = "success", title = "", text = "", footer = "" } = c;
    Swal.fire({ icon, title, text, footer });
  };

  const customModal = async (c) => {
    const {
      title = "",
      icon = "",
      message = "",
      showConfirmButton = true,
      confirmButtonText = "Submit",
    } = c;

    const { value: result } = await Swal.fire({
      title,
      icon,
      html: message,
      focusConfirm: false,
      showConfirmButton,
      confirmButtonText,
      showCancelButton: true,
      willOpen: () => {
        if (c.willOpen !== undefined) {
          c.willOpen();
        }
      },
      didOpen: () => {
        if (c.didOpen !== undefined) {
          c.didOpen();
        }
      },
      preConfirm: () => {
        if (c.preConfirm !== undefined) {
          c.preConfirm();
        }
      },
    });

    if (result) {
      if (result.dismis !== Swal.DismissReason.cancel) {
        if (result.value !== "") {
          if (c.callback !== undefined) {
            c.callback(result);
          }
        } else {
          c.callback(false);
        }
      } else {
        c.callback(false);
      }
    }
  };

  return { toast, alertModal, customModal };
};
