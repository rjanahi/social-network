let isErrorState = false;

function errorPage(errNum) {
  const errorSection = document.getElementById("errorSection");
  const errorContainer = document.getElementById("errorContainer");

  fetch("http://localhost:8080/error/" + errNum, {
    method: "GET",
    credentials: "include",
  })
    .then((res) => {
      return res.text();
    })
    .then((html) => {
      console.log(html);
      // Show the error section based on the error number
      switch (errNum) {
        case 400:
          window.location.href="/error";
          errorContainer.innerHTML =
            "<h1>400 Bad Request</h1><p>Your request could not be understood.</p>";
          isErrorState = true;
          break;
        case 404:
          window.location.href="/error";
          errorContainer.innerHTML =
            "<h1>404 Not Found</h1><p>The resource you are looking for could not be found.</p>";
          isErrorState = true;
          break;
        case 500:
          window.location.href="/error";
          errorContainer.innerHTML =
            "<h1>500 Internal Server Error</h1><p>Something went wrong.</p>";
          isErrorState = true;
          break;
        default:
          window.location.href="/error";
          errorContainer.innerHTML =
            "<h1>404 Not Found</h1><p>The resource you are looking for could not be found.</p>";
          isErrorState = true;
          break;
      }
    })
    .catch((err) => {
      console.error("Failed to fetch error details:", err);
    });
}

window.errorPage = errorPage;
window.isErrorState = isErrorState;