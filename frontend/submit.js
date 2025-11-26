const form = document.querySelector("#offspace");
const editKey = document.querySelector("#editKey")
const preview = document.querySelector("#preview");
const imageElement = document.querySelector("#Photo");
const modalOne = new bootstrap.Modal(document.querySelector("#firstModal"));
const modalTwo = new bootstrap.Modal(document.querySelector("#editModal"));
let adminKey = "";

async function getOffspace() {
    if(text.length !== 36) {
        try {
            const response = await fetch(`http://localhost:3333/?adminKey=${text}`, {
                method: "GET",
            });
        } catch (e) {
            console.error(e);
        }
    } else {
        try {
            let text = editKey.value
            console.log(text);
            const response = await fetch(`http://localhost:3333/get/?editKey=${text}`, {
                method: "GET",
            });
            fillForm(form, await response.json());
            imageElement.removeAttribute("required");
            imageElement.previousElementSibling.innerHTML = "To change the image, upload a new one below  (max 1Mb):";
            modalOne.hide();
            modalTwo.show();
        } catch (e) {
            console.error(e);
        }
    }
}

async function newOffspace() {
    document.querySelector("#preview").src = "preview.svg";
    fillForm(form, {'Name': '', 'Street': '', 'Postcode': '', 'City': '', 'Website': '', 'SocialMedia': ''})
    modalOne.hide();
    imageElement.setAttribute("required", "");
    imageElement.previousElementSibling.innerHTML = "Upload a representative picture of your offspace  (max 1Mb):";
    modalTwo.show()
}

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

editKey.addEventListener("submit", (event) => {
    event.preventDefault();
    getOffspace();
});

// Take over form submission
form.addEventListener("submit", (event) => {
    event.preventDefault();
    const button = document.getElementById("submitButton")
    button.disable();
    sendData();
});

imageElement.addEventListener("change", (event) => {
    const [file] = imageElement.files;
    if (file) {
        preview.src = URL.createObjectURL(file);
    }
})

function fillForm(form, data) {
    Object.keys(data).forEach(key => {
        const field = form.elements.namedItem(key);
        if (!field) return; // skip if form field doesn't exist

        const value = data[key];

        if (field.type === "file") {
            preview.src = data[key];
        } else {
            field.value = value;
        }
    });
}