// Copyright 2016 Documize Inc. <legal@documize.com>. All rights reserved.
//
// This software (Documize Community Edition) is licensed under
// GNU AGPL v3 http://www.gnu.org/licenses/agpl-3.0.en.html
//
// You can operate outside the AGPL restrictions by purchasing
// Documize Enterprise Edition and obtaining a commercial license
// by contacting <sales@documize.com>.
//
// https://documize.com

import Model from 'ember-data/model';
import attr from 'ember-data/attr';
// import { belongsTo, hasMany } from 'ember-data/relationships';

export default Model.extend({
	orgId: attr('string'),
	folderId: attr('string'),
	userId: attr('string'),
	fullname: attr('string'), // client-side usage only, not from API

	spaceView: attr('boolean'),
	spaceManage: attr('boolean'),
	spaceOwner: attr('boolean'),
	documentAdd: attr('boolean'),
	documentEdit: attr('boolean'),
	documentDelete: attr('boolean'),
	documentMove: attr('boolean'),
	documentCopy: attr('boolean'),
	documentTemplate: attr('boolean')
});
