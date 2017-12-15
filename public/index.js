(function(window, document) {

    var supportedFiles = ['image/jpeg', 'image/png'];
    var isOkToUpload = false;

    eStatus = document.getElementById('status');
    eProgress = document.getElementById('progress');
    eProgressBar = document.getElementById('progress_bar');
    eDownloadSrc = document.getElementById('download_link_source');
    eDownloadCmp = document.getElementById('download_link_compressed');
    eDownloadLos = document.getElementById('download_link_lossy');
    eUploadStatus = document.getElementById('file_upload_status');
    eForm = document.getElementById('file_upload');
    eFieldset = document.getElementById('field_upload');

    eForm.addEventListener('submit', onSubmitForm);
    document.getElementById('file').addEventListener('change', onFileChange);

    function onSubmitForm(e) {
        e.preventDefault();
        
        if(!isOkToUpload) {
            return false;
        }
    
        eStatus.classList.remove('hidden');
        eProgressBar.classList.remove('hidden');
        eDownloadSrc.classList.add('hidden');
        eDownloadCmp.classList.add('hidden');
        eDownloadLos.classList.add('hidden');
        eProgress.style.width = '0%';
        eFieldset.disabled = true;
        
        var formData = new FormData();
        var file = document.getElementById('file');
        formData.append('upload[file]', file.files[0]);
        formData.append('upload[name]', file.files[0].name);
        
        eStatus.innerHTML = 'Uploading... 0%';
        
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/upload');
        xhr.addEventListener('load', onRequestComplete, false);
        xhr.upload.addEventListener('load', onUploadComplete, false);
        xhr.upload.addEventListener('progress', onUploadProgress, false);
        xhr.send(formData);
    }

    function onFileChange(e) {
        var files = e.target.files;
        for(var i = 0, len = files.length; i < len; i++) {
            if(supportedFiles.indexOf(files[i].type) === -1) {
                isOkToUpload = false
                break
            }
            if(i == len-1) {
                isOkToUpload = true
            }
        }

        if(isOkToUpload) {
            eUploadStatus.classList.add('hidden')
        } else {
            eUploadStatus.classList.remove('hidden')
        }
    }
    
    function onRequestComplete(e) {
        eFieldset.disabled = false;
        var xhr = e.target
        if(xhr.readyState == 4 && xhr.status == 200) {
            eStatus.innerHTML = 'Finished';
            try {
                handleUploadResponse(JSON.parse(xhr.responseText))
            } catch(e) {
                console.error(e)
            }
        } else {
            eStatus.innerHTML = 'Upload Error. Try again.';
        }
        console.log('Request Completed', e);
    }
    
    function onUploadComplete(e) {
        eStatus.innerHTML = 'Upload Completed. Compressing...';
        eProgressBar.classList.add('hidden');
    }
    
    function onUploadProgress(e) {
        var p = parseFloat(e.loaded / e.total * 100).toFixed(2);
        eStatus.innerHTML = 'Uploading... ' + p + '%';
        eProgress.style.width = p + '%';
    }

    function handleUploadResponse(data) {
        const upload = data.result
        console.log('Upload Response', upload)
        
        var srcSize = upload.source.size;
        var cmpSize = upload.compressed.size;
        
        var diff = Math.abs(srcSize - cmpSize);
        var p = srcSize >= cmpSize ? cmpSize / srcSize : srcSize / cmpSize;
        p = 100 - (p * 100);
        
        var status = 'Original: ' + sizeFormat(srcSize, 2);
        status += '<br><strong>Lossless</strong>: ' + sizeFormat(cmpSize, 2);
        status += ' (' + sizeFormat(diff, 2) + ' -' + p.toFixed(2) + '%)';
        
        eDownloadSrc.href = upload.source.url;
        eDownloadCmp.href = upload.compressed.url;
        eDownloadSrc.classList.remove('hidden');
        eDownloadCmp.classList.remove('hidden');
        
        if(upload.hasOwnProperty('lossy')) {
            var losSize = upload.lossy.size;
            var diffLossy = Math.abs(srcSize - losSize);
            var pLossy = srcSize >= losSize ? losSize / srcSize : srcSize / losSize;
            pLossy = 100 - (pLossy * 100);
            status += '<br><strong>Lossy</strong>: ' + sizeFormat(losSize, 2);
            status += ' (' + sizeFormat(diffLossy, 2) + ' -' + pLossy.toFixed(2) + '%)';
            eDownloadLos.href = upload.lossy.url;
            eDownloadLos.classList.remove('hidden');
        }

        eStatus.innerHTML = status;
    }

    function sizeFormat(size, decimals) {
        var types = ['byte', 'kb', 'mb'];
        i = 0;
        while(size >= 1024 && i <= types.length) {
            size = size / 1024;
            i++
        }
        if(typeof decimals !== 'undefined') {
            size = size.toFixed(decimals)
        }
        return size + types[i];
    }
})(window, document)