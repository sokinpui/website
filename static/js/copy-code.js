function initializeCopyCodeButtons(container) {
    const codeBlocks = container.querySelectorAll('.markdown-body pre');

    codeBlocks.forEach(codeBlock => {
        if (codeBlock.parentNode.classList.contains('code-block-wrapper')) {
            return;
        }

        const wrapper = document.createElement('div');
        wrapper.className = 'code-block-wrapper';

        codeBlock.parentNode.insertBefore(wrapper, codeBlock);
        wrapper.appendChild(codeBlock);

        const copyButton = document.createElement('button');
        copyButton.className = 'copy-code-button';
        copyButton.textContent = 'Copy';
        wrapper.appendChild(copyButton);

        copyButton.addEventListener('click', () => {
            const codeElement = codeBlock.querySelector('code');
            if (navigator.clipboard && codeElement) {
                navigator.clipboard.writeText(codeElement.textContent).then(() => {
                    copyButton.textContent = 'Copied!';
                    setTimeout(() => {
                        copyButton.textContent = 'Copy';
                    }, 2000);
                }).catch(err => {
                    console.error('Failed to copy text: ', err);
                    copyButton.textContent = 'Error';
                });
            }
        });
    });
}

document.addEventListener('DOMContentLoaded', () => {
    initializeCopyCodeButtons(document.body);
});

document.body.addEventListener('htmx:afterSwap', function (event) {
    initializeCopyCodeButtons(event.target);
});
