const form = document.querySelector("#offspace");
const editKey = document.querySelector("#editKey")
const preview = document.querySelector("#preview");
const imageElement = document.querySelector("#photo");
const modalOne = new bootstrap.Modal(document.querySelector("#firstModal"));
const modalTwo = new bootstrap.Modal(document.querySelector("#editModal"));
const modalThree = new bootstrap.Modal(document.querySelector("#messageModal"));
const editcode = new bootstrap.Modal(document.querySelector("#editcodeoutput"));

async function getOffspace() {
    let text = editKey.value
    if(!text) {
        alert("Please enter a valid key.")
    } else if(text.length !== 36) {
        try {
            const response = await fetch(`http://localhost:3333/?adminKey=${text}`, {
                method: "GET",
            });
            offspaceData = await response.json();
            adminKey = text;
            document.getElementById("content").innerHTML = "";
            const frag = document.createDocumentFragment();
            offspaceData.data.forEach(function (v, i) {
                renderCard(frag, v, true, i);
            })
            document.getElementById("content").appendChild(frag);
            modalOne.hide();
        } catch (e) {
            console.error(e);
            alert("No offspace with this edit code could be found.");
        }
    } else {
        try {
            console.log(text);
            const response = await fetch(`http://localhost:3333/get/?editKey=${text}`, {
                method: "GET",
            });
            fillForm(form, await response.json());
            imageElement.removeAttribute("required");
            imageElement.previousElementSibling.innerHTML = "To change the image, upload a new one below  (max 1Mb):";
            const button = document.getElementById("submitButton")
            button.innerHTML = "Update";
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
    const button = document.getElementById("submitButton")
    button.innerHTML = "Submit";
    modalTwo.show();
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
        modalTwo.hide();
        editcode.innerHTML = await response.text();
        modalThree.show();
    } catch (e) {
        console.error(e);
    }
}

async function formDataToJSONWithBase64(formData) {
    const obj = {};

    // First convert regular form fields + base64 files
    for (const [key, value] of formData.entries()) {
        if (value instanceof File) {
            obj[key] = await fileToBase64(value);
        } else {
            obj[key] = value;
        }
    }
    const days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

    const openingTimes = {};

    days.forEach(day => {
        const closed = formData.get(day.toLowerCase() + "_closed") === "on";

        if (closed) {
            openingTimes[day] = { Start: "", End: "" };
        } else {
            openingTimes[day] = {
                Start: formData.get(day.toLowerCase() + "_start") || "",
                End: formData.get(day.toLowerCase() + "_end") || ""
            };
        }
    });

    obj["opening_times"] = openingTimes;

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
    button.disaled = true;
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
        if (key === "opening_times" || key === "openingTimes") return;

        const field = form.elements.namedItem(key);
        if (!field) return;

        const value = data[key];

        if (field.type === "file") {
            // For images
            const preview = document.getElementById("preview");
            if (preview) preview.src = value;
        } else {
            field.value = value;
        }
    });

    const ot = data.opening_times || data.openingTimes;
    if (!ot) return;

    const days = ["mon","tue","wed","thu","fri","sat","sun"];

    days.forEach(day => {
        const d = ot[day];
        const closedField = form.elements.namedItem(`${day}_closed`);
        const startField  = form.elements.namedItem(`${day}_start`);
        const endField    = form.elements.namedItem(`${day}_end`);

        if (!closedField || !startField || !endField) return;

        if (!d || d.closed) {
            // It's closed
            closedField.checked = true;

            // Disable selects
            startField.disabled = true;
            endField.disabled = true;

            // Clear values
            startField.value = "";
            endField.value = "";
        } else {
            closedField.checked = false;

            // Enable selects
            startField.disabled = false;
            endField.disabled = false;

            // Assign values (if present)
            if (d.start) startField.value = d.start;
            if (d.end)   endField.value   = d.end;
        }
    });
}


function editOffspace(data) {
    fillForm(form, data);
    imageElement.removeAttribute("required");
    imageElement.previousElementSibling.innerHTML = "To change the image, upload a new one below  (max 1Mb):";
    const button = document.getElementById("submitButton")
    button.innerHTML = "Update";
    modalOne.hide();
    modalTwo.show();
}

document.getElementById('copyBtn').addEventListener('click', async () => {
    try {
        await navigator.clipboard.writeText(editcode.value);
        alert('Edit code copied to clipboard!');
    } catch (err) {
        console.error('Failed to copy edit code: ', err);
    }
});