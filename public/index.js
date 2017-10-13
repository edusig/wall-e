(function(window, document) {

    var supportedFiles = ['image/jpeg', 'image/png'];
    var isOkToUpload = false;

    eStatus = document.getElementById('status');
    eProgress = document.getElementById('progress');
    eProgressBar = document.getElementById('progress_bar');
    eDownload = document.getElementById('download_link');
    eUploadStatus = document.getElementById('file_upload_status');

    document.getElementById('file_upload').addEventListener('submit', onSubmitForm);
    document.getElementById('file').addEventListener('change', onFileChange);

    function onSubmitForm(e) {
        e.preventDefault();
        
        if(!isOkToUpload) {
            return false;
        }
    
        eStatus.classList.remove('hidden');
        eProgress.classList.remove('hidden');
        eDownload.classList.add('hidden');
        eProgress.style.width = '0%';
        
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
        var xhr = e.target
        if(xhr.readyState == 4 && xhr.status == 200) {
            eStatus.innerHTML = 'Finished';
            eDownload.classList.remove('hidden');
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
        eProgress.classList.add('hidden');
    }
    
    function onUploadProgress(e) {
        var p = parseFloat(e.loaded / e.total * 100).toFixed(2);
        eStatus.innerHTML = 'Uploading... ' + p + '%';
        eProgress.style.width = p + '%';
    }

    function handleUploadResponse(upload) {
        console.log('Upload Response', upload)
        var srcSize = upload.source.size;
        var cmpSize = upload.compressed.size;
        var diff = Math.abs(srcSize - cmpSize) / 1024;
        var p = srcSize >= cmpSize ? cmpSize / srcSize : srcSize / cmpSize;
        p = 100 - (p * 100);
        var status = 'Original File Size: ' + sizeFormat(srcSize, 2);
        status += '<br>Compressed File Size: ' + sizeFormat(cmpSize, 2);
        status += '<br>Saved: ' + sizeFormat(diff, 2) + ' (' + p.toFixed(2) + '%)';
        console.log(status);
        eStatus.innerHTML = status;
    }

    function sizeFormat(size, decimals) {
        var types = ['byte', 'kb', 'mb'];
        i = 0;
        while(size >= 1024 && i < 2) {
            size = size / 1024;
            i++
        }
        if(typeof decimals !== 'undefined') {
            size = size.toFixed(decimals)
        }
        return size + types[i];
    }
})(window, document)