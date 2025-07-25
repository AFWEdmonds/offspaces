const form = document.querySelector("#offspace");

async function sendData() {
    // Associate the FormData object with the form element
    const formData = new FormData(form);
    try {
        let blah = await formDataToJSONWithBase64(formData);
        const response = await fetch("http://localhost:3333/create/", {
            method: "POST",
            // Set the FormData instance as the request body
            body: blah,
        });
        console.log(await response.json());
    } catch (e) {
        console.error(e);
    }
}

async function formDataToJSONWithBase64(formData) {
    const obj = {};
    for (const [key, value] of formData.entries()) {
        if (value instanceof File) {
            obj[key] = await fileToBase64(value);
        } else {
            obj[key] = value;
        }
    }
    return JSON.stringify(obj);
}

function fileToBase64(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result); // includes the data:image/... prefix
        reader.onerror = reject;
        reader.readAsDataURL(file);
    });
}

// Take over form submission
form.addEventListener("submit", (event) => {
    event.preventDefault();
    const button = document.getElementById("submitButton")
    button.disable();
    sendData();
});