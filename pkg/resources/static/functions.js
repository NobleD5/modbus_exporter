// Namespace.
var modbusexporter = {};

modbusexporter.labels = {};
modbusexporter.panel = null;

modbusexporter.switchToMetrics = function(){
    $('#metrics-div').show();
    $('#status-div').hide();
    $('#config-div').hide();
    $('#metrics-li').addClass('active');
    $('#status-li').removeClass('active');
    $('#config-li').removeClass('active');
}

modbusexporter.switchToStatus = function(){
    $('#metrics-div').hide();
    $('#status-div').show();
    $('#config-div').hide();
    $('#metrics-li').removeClass('active');
    $('#status-li').addClass('active');
    $('#config-li').removeClass('active');
}

modbusexporter.switchToConfig = function(){
    $('#metrics-div').hide();
    $('#status-div').hide();
    $('#config-div').show();
    $('#metrics-li').removeClass('active');
    $('#status-li').removeClass('active');
    $('#config-li').addClass('active');
}

modbusexporter.showDelModal = function(labels, labelsEncoded, panelID, event){
    event.stopPropagation(); // Don't trigger accordion collapse.
    modbusexporter.labels = labelsEncoded;
    modbusexporter.panel = $('#' + panelID).parent();

    var components = [];
    for (var ln in labels) {
	components.push(ln + '="' + labels[ln] + '"')
    }

    $('#del-modal-msg').text(
	'Do you really want to delete all metrics of group {' + components.join(', ') + '}?'
    );
    $('#del-modal').modal('show');
}

modbusexporter.showDelAllModal = function(){
    if (!$('button#del-all').hasClass('disabled')) {
        $('#del-modal-all-msg').text(
            'Do you really want to delete all metrics from all metric groups?'
        );
        $('#del-all-modal').modal('show');
    }
}

modbusexporter.deleteGroup = function(){
    var pathElements = [];
    for (var ln in modbusexporter.labels) {
	if (ln != 'job') {
	    pathElements.push(encodeURIComponent(ln+'@base64'));
	    pathElements.push(encodeURIComponent(modbusexporter.labels[ln]));
	}
    }
    var groupPath = pathElements.join('/');
    if (groupPath.length > 0) {
	groupPath = '/' + groupPath;
    }

    $.ajax({
	type: 'DELETE',
	url: 'metrics/job@base64/' + encodeURIComponent(modbusexporter.labels['job']) + groupPath,
	success: function(data, textStatus, jqXHR) {
	    modbusexporter.panel.remove();
        modbusexporter.decreaseDelAllCounter();
	    $('#del-modal').modal('hide');
	},
	error: function(jqXHR, textStatus, error) {
	    alert('Deleting metric group failed: ' + error);
	}
    });
}

modbusexporter.deleteAllGroup = function(){
    $.ajax({
        type: 'PUT',
        url: 'api/v1/admin/wipe',
        success: function(data, textStatus, jqXHR) {
            $('div').each(function() {
                id = $(this).attr("id");
                if (typeof id != 'undefined' && id.match(/^group-panel-[0-9]{1,}$/)) {
                    $(this).parent().remove();
                }
            });
            modbusexporter.setDelAllCounter(0);
            $('#del-all-modal').modal('hide');
        },
        error: function(jqXHR, textStatus, error) {
            alert('Deleting all metric groups failed: ' + error);
        }
    });
}

modbusexporter.decreaseDelAllCounter = function(){
    var counter = parseInt($('span#del-all-counter').text());
    modbusexporter.setDelAllCounter(--counter);
}

modbusexporter.setDelAllCounter = function(n){
    $('span#del-all-counter').text(n);
    if (n <= 0) {
        modbusexporter.disableDelAllGroupButton();
        return;
    }
    modbusexporter.enableDelAllGroupButton();
}

modbusexporter.enableDelAllGroupButton = function(){
    $('button#del-all').removeClass('disabled');
}

modbusexporter.disableDelAllGroupButton = function(){
    $('button#del-all').addClass('disabled');
}

$(function () {
    $('div.collapse').on('show.bs.collapse', function (event) {
	$(this).prev().find('span.toggle-icon')
	    .removeClass('glyphicon-collapse-down')
	    .addClass('glyphicon-collapse-up');
	event.stopPropagation();
    })
    $('div.collapse').on('hide.bs.collapse', function (event) {
	$(this).prev().find('span.toggle-icon')
	    .removeClass('glyphicon-collapse-up')
	    .addClass('glyphicon-collapse-down');
	event.stopPropagation();
    })
})
