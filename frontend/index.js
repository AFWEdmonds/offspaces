const content = document.getElementById("#content");

async function onInit() {
    try {
        const response = await fetch("http://localhost:3333/", {
            method: "GET",
        });
        const data = await response.json();
        render(data);
    } catch (e) {
        console.error(e);
    }
}

function render(data) {
    const frag = document.createDocumentFragment();

    data.forEach(function (v, i) {
        const card = document.createElement('div');
        card.className = 'card my-4';

        const img = document.createElement('img');
        img.className = 'card-img-top';
        img.alt = 'The offspace';
        img.src = v.Photo;
        card.appendChild(img);

        const cardBody = document.createElement('div');
        cardBody.className = 'card-body';

        const cardTitle = document.createElement('h5');
        cardTitle.className = 'card-title';
        cardTitle.innerHTML = v.Name;
        cardBody.appendChild(cardTitle);

        const cardText = document.createElement('p');
        cardText.className = 'card-text';
        cardText.innerHTML = `${v.Street}<br>${v.Postcode} ${v.City}`;
        cardBody.appendChild(cardText);

        const visitWebsite = document.createElement('a');
        visitWebsite.href = v.Website;
        visitWebsite.className = 'card-link';
        visitWebsite.innerHTML = "Website";
        cardBody.appendChild(visitWebsite);

        const visitSM = document.createElement('a');
        visitSM.href = v.SocialMedia;
        visitSM.className = 'card-link';
        visitSM.innerHTML = "Social media";
        cardBody.appendChild(visitSM);

        card.appendChild(cardBody);
        frag.appendChild(card);
    });

    document.getElementById("content").appendChild(frag);
}

onInit();