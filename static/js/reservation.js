// function for disabling form submissions if there are invalid fields
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

  const html = `
  <form
    id="availability-form"
    action=""
    method="POST"
    class="needs-validation"
    novalidate
  >
    <div class="form-row">
      <div class="col">
        <div class="form-row" id="reservation-dates-modal">
          <div class="col mb-3">
            <input
              type="text"
              class="form-control"
              name="start"
              id="start"
              placeholder="Arrival date"
              autocomplete="off"
              required
              disabled
            />
          </div>
          <div class="col">
            <input
              type="text"
              class="form-control"
              name="end"
              id="end"
              placeholder="Depature date"
              autocomplete="off"
              required
              disabled
            />
          </div>
        </div>
      </div>
    </div>
  </form>
`;

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
      const { title = "", html = "" } = c;

      const { value: formValues } = await Swal.fire({
        title,
        html,
        focusConfirm: false,
        confirmButtonText: "Submit",
        showCancelButton: true,
        willOpen: () => {
          const elem = document.getElementById("reservation-dates-modal");
          const datepicker = new DateRangePicker(elem, {
            format: "dd-mm-yyyy",
            showOnFocus: true,
          });
        },
        didOpen: () => {
          document.getElementById("start").removeAttribute("disabled");
          document.getElementById("end").removeAttribute("disabled");
        },
        preConfirm: () => {
          return [
            document.getElementById("start").value,
            document.getElementById("end").value,
          ];
        },
      });

      if (formValues) {
        Swal.fire(JSON.stringify(formValues));
      }
    };

    return { toast, alertModal, customModal };
  };